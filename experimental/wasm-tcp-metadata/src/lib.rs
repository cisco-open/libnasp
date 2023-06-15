use byteorder::{BigEndian, ByteOrder, LittleEndian};
use flatbuffers::{ForwardsUOffset, Vector, WIPOffset};
use log::debug;
use node_info_generated::wasm::common::{
    root_as_flat_node, FlatNode, FlatNodeBuilder, KeyVal, KeyValArgs,
};
use protobuf::{
    well_known_types::{any::Any, struct_::Struct, struct_::Value},
    CodedOutputStream, Message,
};
use proxy_wasm::{
    hostcalls,
    traits::{Context, RootContext, StreamContext},
    types::{Action, ContextType, LogLevel},
};
use std::str;
extern crate flatbuffers;

#[allow(dead_code, unused_imports)]
#[path = "./node_info_generated.rs"]
mod node_info_generated;

const TRAFFIC_DIRECTION_INBOUND: i64 = 1;
const TRAFFIC_DIRECTION_OUTBOUND: i64 = 2;
const MAGIC_NUMBER: u32 = 1025705063;
const MX_ALPN: &str = "istio-peer-exchange";
const PEER_ID_METADATA_KEY: &str = "x-envoy-peer-metadata-id";
const PEER_METADATA_KEY: &str = "x-envoy-peer-metadata";

#[no_mangle]
pub fn _start() {
    proxy_wasm::set_log_level(LogLevel::Trace);
    proxy_wasm::set_root_context(|context_id| -> Box<dyn RootContext> {
        Box::new(MetadataFilter {
            context_id,
            metadata_value: String::new(),
            node_id: String::new(),
            alpn: String::new(),
            direction: 0,
            metadata_id_received: false,
            metadata_received: false,
            metadata_exchanged: false,
            metadata_sent: false,
        })
    });
}

struct MetadataFilter {
    context_id: u32,
    metadata_value: String,
    node_id: String,
    alpn: String,
    direction: i64,
    metadata_id_received: bool,
    metadata_received: bool,
    metadata_exchanged: bool,
    metadata_sent: bool,
}

impl MetadataFilter {
    fn update_metadata_value(&mut self) {
        let node_info = extract_local_node_flat_buffer();
        let root_flatnode = root_as_flat_node(&node_info).unwrap();
        let metadata = extract_struct_from_node_flat_buffer(&root_flatnode, &self.node_id);
        let mut metadata_bytes = Vec::<u8>::new();
        serialize_to_string_deterministic(metadata, &mut metadata_bytes);
        self.metadata_value = base64::encode(metadata_bytes);
    }
}

impl Context for MetadataFilter {}

impl RootContext for MetadataFilter {
    fn on_configure(&mut self, _plugin_configuration_size: usize) -> bool {
        if let Some(node_id) = self.get_property(vec!["node", "id"]) {
            self.node_id = str::from_utf8(&node_id).unwrap().to_string();
        } else {
            debug!("[context: {}] cannot get node ID", self.context_id);
            return false;
        }

        self.update_metadata_value();
        debug!(
            "[context: {}] metadata_value_ : {}, node: {}",
            self.context_id, self.metadata_value, self.node_id
        );
        true
    }

    fn create_stream_context(&self, _context_id: u32) -> Option<Box<dyn StreamContext>> {
        Some(Box::new(MetadataFilter {
            context_id: _context_id,
            metadata_value: self.metadata_value.to_string(),
            node_id: self.node_id.to_string(),
            alpn: self.alpn.to_string(),
            direction: self.direction,
            metadata_id_received: self.metadata_id_received,
            metadata_received: self.metadata_received,
            metadata_exchanged: self.metadata_exchanged,
            metadata_sent: self.metadata_sent,
        }))
    }

    fn get_type(&self) -> Option<ContextType> {
        Some(ContextType::StreamContext)
    }
}

impl MetadataFilter {
    fn set_direction(&mut self) {
        if self.direction == 0 {
            let direction_buf = self.get_property(vec!["listener_direction"]).unwrap();
            self.direction = LittleEndian::read_i64(&direction_buf);
        }
    }

    fn set_alpn(&mut self) {
        if self.alpn == "" {
            let alpn_buf = self
                .get_property(vec!["upstream", "negotiated_protocol"])
                .unwrap_or(Vec::<u8>::new());
            self.alpn = str::from_utf8(&alpn_buf).unwrap().to_string();
        }
    }
}

impl StreamContext for MetadataFilter {
    fn on_downstream_data(&mut self, _data_size: usize, _end_of_stream: bool) -> Action {
        debug!(
            "[context: {}] on downstream data ({} bytes)",
            self.context_id, _data_size
        );

        if self.metadata_exchanged {
            debug!("[context: {}] metadata already exchanged", self.context_id);
            return Action::Continue;
        }

        self.set_direction();
        self.set_alpn();

        debug!(
            "[context: {}] on_downstream_data ALPN: [{}], direction: [{}]",
            self.context_id, self.alpn, self.direction
        );

        match self.direction {
            // send local metadata by prepending it into the downstream buffer
            TRAFFIC_DIRECTION_OUTBOUND => {
                if self.alpn != MX_ALPN || self.metadata_sent {
                    return Action::Continue;
                }

                let mut buffer = Vec::<u8>::new();

                let mut u32buf = [0; 4];
                BigEndian::write_u32(&mut u32buf, MAGIC_NUMBER);
                buffer.extend_from_slice(&u32buf);

                let metadata_bytes = base64::decode(&self.metadata_value).unwrap();
                BigEndian::write_u32(&mut u32buf, metadata_bytes.len() as u32);
                buffer.extend_from_slice(&u32buf);
                buffer.extend(&metadata_bytes);

                debug!(
                    "[context: {}] prepend metadata value ({} bytes) to downstream buffer",
                    self.context_id,
                    buffer.len()
                );
                self.set_downstream_data(0, 0, &buffer);
                self.metadata_sent = true;

                Action::Continue
            }
            // parse incoming metadata
            TRAFFIC_DIRECTION_INBOUND => {
                // check if we have at least 4 byte
                let u32buf = self.get_downstream_data(0, 4).unwrap();
                if u32buf.len() < 4 {
                    return Action::Continue;
                }

                debug!("[context: {}] parse incoming metadata", self.context_id);

                let magic_number = BigEndian::read_u32(&u32buf);
                if magic_number == MAGIC_NUMBER {
                    debug!("[context: {}] Magic number found!", self.context_id);
                    let length_buf = self.get_downstream_data(4, 4).unwrap();
                    let length = BigEndian::read_u32(&length_buf);
                    let data = self.get_downstream_data(8, length as usize).unwrap();

                    //TODO handle error
                    let md2 = Any::parse_from_bytes(&data).unwrap();
                    let md = Struct::parse_from_bytes(md2.value.as_slice()).unwrap();

                    if let Some(downstream_metadata_id) = md.fields.get(PEER_ID_METADATA_KEY) {
                        self.set_property(
                            vec!["downstream_peer_id"],
                            Some(downstream_metadata_id.string_value().as_bytes()),
                        );
                        self.metadata_id_received = true;
                    }

                    if let Some(downstream_metadata) = md.fields.get(PEER_METADATA_KEY) {
                        let fb = extract_node_flat_buffer_from_struct(
                            downstream_metadata.struct_value(),
                        );
                        self.set_property(vec!["downstream_peer"], Some(&fb));
                        self.metadata_received = true;
                    }

                    debug!(
                        "[context: {}] downstream peer node info set",
                        self.context_id
                    );

                    // the incoming metadata should not be sent to the client
                    // so leave only the rest in the downstream buffer
                    let remainder = self
                        .get_downstream_data(length as usize + 8, _data_size)
                        .unwrap();
                    debug!(
                        "[context: {}] remainder in the downstream buf {} bytes",
                        self.context_id,
                        remainder.len()
                    );
                    self.set_downstream_data(0, usize::MAX, &remainder);

                    // send local metadata back by prepending it into the upstream buffer
                    if self.metadata_received && self.metadata_id_received && !self.metadata_sent {
                        let mut buffer = Vec::<u8>::new();

                        let mut u32buf = [0; 4];
                        BigEndian::write_u32(&mut u32buf, magic_number);
                        buffer.extend_from_slice(&u32buf);

                        let metadata_bytes = base64::decode(&self.metadata_value).unwrap();
                        BigEndian::write_u32(&mut u32buf, metadata_bytes.len() as u32);
                        buffer.extend_from_slice(&u32buf);
                        buffer.extend(&metadata_bytes);

                        self.set_upstream_data(0, 0, &buffer);
                        debug!("[context: {}] upstream data set", self.context_id);
                        self.metadata_sent = true;
                    }
                }

                Action::Continue
            }
            _ => Action::Continue,
        }
    }

    fn on_upstream_data(&mut self, _data_size: usize, _end_of_stream: bool) -> Action {
        debug!(
            "[context: {}] on upstream data {}",
            self.context_id, _data_size
        );

        if self.metadata_exchanged {
            debug!("[context: {}] metadata already exchanged", self.context_id);
            return Action::Continue;
        }

        self.set_direction();
        self.set_alpn();

        debug!(
            "[context: {}] on_upstream_data ALPN: [{}], direction: [{}]",
            self.context_id, self.alpn, self.direction
        );

        match self.direction {
            // parse incoming metadata
            TRAFFIC_DIRECTION_OUTBOUND => {
                // check if we have at least 4 byte
                let u32buf = self.get_upstream_data(0, 4).unwrap();
                if u32buf.len() < 4 {
                    return Action::Continue;
                }

                let magic_number = BigEndian::read_u32(&u32buf);
                if magic_number == MAGIC_NUMBER {
                    debug!("[context: {}] magic number found!", self.context_id);
                    let length_buf = self.get_upstream_data(4, 4).unwrap();
                    let length = BigEndian::read_u32(&length_buf);
                    let data = self.get_upstream_data(8, length as usize).unwrap();

                    //TODO handle error
                    let md2 = Any::parse_from_bytes(&data).unwrap();
                    let md = Struct::parse_from_bytes(md2.value.as_slice()).unwrap();

                    if let Some(upstream_metadata_id) = md.fields.get(PEER_ID_METADATA_KEY) {
                        self.set_property(
                            vec!["upstream_peer_id"],
                            Some(upstream_metadata_id.string_value().as_bytes()),
                        );
                    }
                    if let Some(upstream_metadata) = md.fields.get(PEER_METADATA_KEY) {
                        let fb =
                            extract_node_flat_buffer_from_struct(upstream_metadata.struct_value());
                        self.set_property(vec!["upstream_peer"], Some(&fb));
                    }

                    debug!("[context: {}] upstream peer node set", self.context_id);

                    // the incoming metadata should not be sent to the client
                    // so leave only the rest in the downstream buffer
                    let remainder = self
                        .get_upstream_data(length as usize + 8, _data_size)
                        .unwrap();

                    debug!(
                        "[context: {}] remainder in the upstream buf {} bytes",
                        self.context_id,
                        remainder.len()
                    );

                    self.set_upstream_data(0, usize::MAX, &remainder);

                    self.metadata_exchanged = true;
                }

                Action::Continue
            }
            // send local metadata by prepending it into the downstream buffer
            TRAFFIC_DIRECTION_INBOUND => {
                if !self.metadata_received || !self.metadata_id_received || self.metadata_sent {
                    return Action::Continue;
                }

                let mut buffer = Vec::<u8>::new();

                let mut u32buf = [0; 4];
                BigEndian::write_u32(&mut u32buf, MAGIC_NUMBER);
                buffer.extend_from_slice(&u32buf);

                let metadata_bytes = base64::decode(&self.metadata_value).unwrap();
                BigEndian::write_u32(&mut u32buf, metadata_bytes.len() as u32);
                buffer.extend_from_slice(&u32buf);
                buffer.extend(&metadata_bytes);

                self.set_downstream_data(0, 0, &buffer);

                self.metadata_exchanged = true;
                self.metadata_sent = true;

                Action::Continue
            }
            _ => Action::Continue,
        }
    }
}

fn extract_node_flat_buffer_from_struct(_struct: &Struct) -> Vec<u8> {
    let mut builder = flatbuffers::FlatBufferBuilder::new();

    let mut name: Option<WIPOffset<&str>> = None;
    if let Some(value) = _struct.fields.get("NAME") {
        name = Some(builder.create_string(value.string_value()));
    }

    let mut namespace: Option<WIPOffset<&str>> = None;
    if let Some(value) = _struct.fields.get("NAMESPACE") {
        namespace = Some(builder.create_string(value.string_value()));
    }

    let mut owner: Option<WIPOffset<&str>> = None;
    if let Some(value) = _struct.fields.get("OWNER") {
        owner = Some(builder.create_string(value.string_value()));
    }

    let mut workload_name: Option<WIPOffset<&str>> = None;
    if let Some(value) = _struct.fields.get("WORKLOAD_NAME") {
        workload_name = Some(builder.create_string(value.string_value()));
    }

    let mut istio_version: Option<WIPOffset<&str>> = None;
    if let Some(value) = _struct.fields.get("ISTIO_VERSION") {
        istio_version = Some(builder.create_string(value.string_value()));
    }

    let mut mesh_id: Option<WIPOffset<&str>> = None;
    if let Some(value) = _struct.fields.get("MESH_ID") {
        mesh_id = Some(builder.create_string(value.string_value()));
    }

    let mut cluster_id: Option<WIPOffset<&str>> = None;
    if let Some(value) = _struct.fields.get("CLUSTER_ID") {
        cluster_id = Some(builder.create_string(value.string_value()));
    }

    let mut _app_containers: Vec<flatbuffers::WIPOffset<&str>> = Vec::new();
    let mut app_containers_offset: Option<WIPOffset<Vector<ForwardsUOffset<&str>>>> = None;
    if let Some(value) = _struct.fields.get("APP_CONTAINERS") {
        let containers: Vec<&str> = value.string_value().split_terminator(",").collect();
        for container in containers {
            _app_containers.push(builder.create_string(container))
        }
        app_containers_offset = Some(builder.create_vector(&_app_containers));
    }

    // labels needs to be in a sorted vector
    let mut sorted_labels = std::collections::BTreeMap::new();
    let mut labels: Vec<flatbuffers::WIPOffset<KeyVal>> = Vec::new();
    let mut labels_offset: Option<WIPOffset<Vector<ForwardsUOffset<KeyVal>>>> = None;
    if let Some(value) = _struct.fields.get("LABELS") {
        for field in value.struct_value().fields.iter() {
            sorted_labels.insert(field.0.to_string(), field.1.string_value().to_string());
        }
        for field in sorted_labels {
            let key = builder.create_string(field.0.as_str());
            let value = builder.create_string(field.1.as_str());
            labels.push(KeyVal::create(
                &mut builder,
                &KeyValArgs {
                    key: Some(key),
                    value: Some(value),
                },
            ));
        }
        labels_offset = Some(builder.create_vector(&labels));
    }

    let mut sorted_platform_metadata = std::collections::BTreeMap::new();
    let mut platform_metadata: Vec<flatbuffers::WIPOffset<KeyVal>> = Vec::new();
    let mut platform_metadata_offset: Option<WIPOffset<Vector<ForwardsUOffset<KeyVal>>>> = None;
    if let Some(value) = _struct.fields.get("PLATFORM_METADATA") {
        for field in value.struct_value().fields.iter() {
            sorted_platform_metadata
                .insert(field.0.to_string(), field.1.string_value().to_string());
        }
        for field in sorted_platform_metadata {
            let key = builder.create_string(field.0.as_str());
            let value = builder.create_string(field.1.as_str());
            platform_metadata.push(KeyVal::create(
                &mut builder,
                &KeyValArgs {
                    key: Some(key),
                    value: Some(value),
                },
            ));
        }
        platform_metadata_offset = Some(builder.create_vector(&platform_metadata));
    }

    let mut ip_addrs: Vec<flatbuffers::WIPOffset<&str>> = Vec::new();
    let mut ip_addrs_offset: Option<WIPOffset<Vector<ForwardsUOffset<&str>>>> = None;
    if let Some(value) = _struct.fields.get("INSTANCE_IPS") {
        let instance_ips: Vec<&str> = value.string_value().split_terminator(",").collect();
        for instance_ip in instance_ips {
            ip_addrs.push(builder.create_string(instance_ip))
        }
        ip_addrs_offset = Some(builder.create_vector(&ip_addrs));
    }

    let mut node = FlatNodeBuilder::new(&mut builder);

    if let Some(n) = name {
        node.add_name(n);
    }
    if let Some(v) = namespace {
        node.add_namespace(v);
    }
    if let Some(v) = owner {
        node.add_owner(v);
    }
    if let Some(v) = workload_name {
        node.add_workload_name(v);
    }
    if let Some(v) = istio_version {
        node.add_istio_version(v);
    }
    if let Some(v) = mesh_id {
        node.add_mesh_id(v);
    }
    if let Some(v) = cluster_id {
        node.add_cluster_id(v);
    }
    if let Some(v) = app_containers_offset {
        node.add_app_containers(v);
    }
    if let Some(v) = labels_offset {
        node.add_labels(v);
    }
    if let Some(v) = platform_metadata_offset {
        node.add_platform_metadata(v);
    }
    if let Some(v) = ip_addrs_offset {
        node.add_instance_ips(v);
    }

    let data = node.finish();
    builder.finish(data, None);
    let buf = builder.finished_data().to_vec();

    buf
}

fn extract_struct_from_node_flat_buffer(
    node: &FlatNode,
    node_id: &str,
) -> protobuf::well_known_types::any::Any {
    let mut peer_metadata = protobuf::well_known_types::struct_::Struct::new();
    if let Some(value) = node.name() {
        let mut temp_value = Value::new();
        temp_value.set_string_value(value.to_string());
        peer_metadata.fields.insert("NAME".to_string(), temp_value);
    }
    if let Some(value) = node.namespace() {
        let mut temp_value = Value::new();
        temp_value.set_string_value(value.to_string());
        peer_metadata
            .fields
            .insert("NAMESPACE".to_string(), temp_value);
    }
    if let Some(value) = node.owner() {
        let mut temp_value = Value::new();
        temp_value.set_string_value(value.to_string());
        peer_metadata.fields.insert("OWNER".to_string(), temp_value);
    }
    if let Some(value) = node.workload_name() {
        let mut temp_value = Value::new();
        temp_value.set_string_value(value.to_string());
        peer_metadata
            .fields
            .insert("WORKLOAD_NAME".to_string(), temp_value);
    }
    if let Some(value) = node.istio_version() {
        let mut temp_value = Value::new();
        temp_value.set_string_value(value.to_string());
        peer_metadata
            .fields
            .insert("ISTIO_VERSION".to_string(), temp_value);
    }
    if let Some(value) = node.mesh_id() {
        let mut temp_value = Value::new();
        temp_value.set_string_value(value.to_string());
        peer_metadata
            .fields
            .insert("MESH_ID".to_string(), temp_value);
    }
    if let Some(value) = node.cluster_id() {
        let mut temp_value = Value::new();
        temp_value.set_string_value(value.to_string());
        peer_metadata
            .fields
            .insert("CLUSTER_ID".to_string(), temp_value);
    }
    if let Some(value) = node.labels() {
        let mut extracted_struct = protobuf::well_known_types::struct_::Struct::new();
        for label_flatbuff in value {
            let mut flatbuff_value = Value::new();
            flatbuff_value.set_string_value(label_flatbuff.value().unwrap().to_string());
            extracted_struct
                .fields
                .insert(label_flatbuff.key().to_string(), flatbuff_value);
        }
        let mut temp_value = Value::new();
        temp_value.set_struct_value(extracted_struct);
        peer_metadata
            .fields
            .insert("LABELS".to_string(), temp_value);
    }
    if let Some(value) = node.platform_metadata() {
        let mut extracted_struct = protobuf::well_known_types::struct_::Struct::new();
        for platform_metadata_flatbuff in value {
            let mut flatbuff_value = Value::new();
            flatbuff_value
                .set_string_value(platform_metadata_flatbuff.value().unwrap().to_string());
            extracted_struct
                .fields
                .insert(platform_metadata_flatbuff.key().to_string(), flatbuff_value);
        }
        let mut temp_value = Value::new();
        temp_value.set_struct_value(extracted_struct);
        peer_metadata
            .fields
            .insert("PLATFORM_METADATA".to_string(), temp_value);
    }
    if let Some(value) = node.app_containers() {
        let mut temp_value = Value::new();
        temp_value.set_string_value(
            value
                .iter()
                .map(|x| x.to_string())
                .collect::<Vec<String>>()
                .join(","),
        );
        peer_metadata
            .fields
            .insert("APP_CONTAINERS".to_string(), temp_value);
    }
    if let Some(value) = node.instance_ips() {
        let mut temp_value = Value::new();
        temp_value.set_string_value(
            value
                .iter()
                .map(|x| x.to_string())
                .collect::<Vec<String>>()
                .join(","),
        );
        peer_metadata
            .fields
            .insert("INSTANCE_IPS".to_string(), temp_value);
    }
    //Prepare the whole metadata struct enhanced with node_id
    let mut metadata = protobuf::well_known_types::struct_::Struct::new();
    let mut metadata_value = Value::new();
    metadata_value.set_struct_value(peer_metadata);
    metadata
        .fields
        .insert(PEER_METADATA_KEY.to_string(), metadata_value);
    let mut metadata_id_value = Value::new();
    metadata_id_value.set_string_value(node_id.to_string());
    metadata
        .fields
        .insert(PEER_ID_METADATA_KEY.to_string(), metadata_id_value);
    let metadata_any = protobuf::well_known_types::any::Any::pack(&metadata).unwrap();
    metadata_any
}

fn extract_local_node_flat_buffer() -> Vec<u8> {
    let mut builder = flatbuffers::FlatBufferBuilder::new();

    let mut labels: Vec<flatbuffers::WIPOffset<KeyVal>> = Vec::new();
    let mut platform_metadata: Vec<flatbuffers::WIPOffset<KeyVal>> = Vec::new();

    let mut app_containers: Vec<flatbuffers::WIPOffset<&str>> = Vec::new();
    let mut ip_addrs: Vec<flatbuffers::WIPOffset<&str>> = Vec::new();

    let mut value = hostcalls::get_property(vec!["node", "metadata", "NAME"])
        .unwrap()
        .unwrap();
    let name = builder.create_string(str::from_utf8(&value).unwrap());
    value = hostcalls::get_property(vec!["node", "metadata", "NAMESPACE"])
        .unwrap()
        .unwrap();
    let namespace_ = builder.create_string(str::from_utf8(&value).unwrap());
    value = hostcalls::get_property(vec!["node", "metadata", "OWNER"])
        .unwrap()
        .unwrap();
    let owner = builder.create_string(str::from_utf8(&value).unwrap());
    value = hostcalls::get_property(vec!["node", "metadata", "WORKLOAD_NAME"])
        .unwrap()
        .unwrap();
    let workload_name = builder.create_string(str::from_utf8(&value).unwrap());
    value = hostcalls::get_property(vec!["node", "metadata", "ISTIO_VERSION"])
        .unwrap()
        .unwrap();
    let istio_version = builder.create_string(str::from_utf8(&value).unwrap());
    value = hostcalls::get_property(vec!["node", "metadata", "MESH_ID"])
        .unwrap()
        .unwrap();
    let mesh_id = builder.create_string(str::from_utf8(&value).unwrap());
    value = hostcalls::get_property(vec!["node", "metadata", "CLUSTER_ID"])
        .unwrap()
        .unwrap();
    let cluster_id = builder.create_string(str::from_utf8(&value).unwrap());

    if let Some(value) = hostcalls::get_property(vec!["node", "metadata", "LABELS"]).unwrap() {
        let labels_map = deserialize_map(&value);
        for elment in labels_map {
            let key = builder.create_string(elment.0.as_str());
            let value = builder.create_string(elment.1.as_str());
            labels.push(KeyVal::create(
                &mut builder,
                &KeyValArgs {
                    key: Some(key),
                    value: Some(value),
                },
            ));
        }
    }
    if let Some(value) =
        hostcalls::get_property(vec!["node", "metadata", "PLATFORM_METADATA"]).unwrap()
    {
        let metadata_map = deserialize_map(&value);
        for elment in metadata_map {
            let key = builder.create_string(elment.0.as_str());
            let value = builder.create_string(elment.1.as_str());
            platform_metadata.push(KeyVal::create(
                &mut builder,
                &KeyValArgs {
                    key: Some(key),
                    value: Some(value),
                },
            ));
        }
    }

    if let Some(value) =
        hostcalls::get_property(vec!["node", "metadata", "APP_CONTAINERS"]).unwrap()
    {
        let containers: Vec<&str> = str::from_utf8(&value)
            .unwrap()
            .split_terminator(",")
            .collect();
        for container in containers {
            app_containers.push(builder.create_string(container))
        }
    }
    if let Some(value) = hostcalls::get_property(vec!["node", "metadata", "INSTANCE_IPS"]).unwrap()
    {
        let ips: Vec<&str> = str::from_utf8(&value)
            .unwrap()
            .split_terminator(",")
            .collect();
        for ip in ips {
            ip_addrs.push(builder.create_string(ip))
        }
    }

    let app_containers_offset = builder.create_vector(&app_containers);
    let ip_addrs_offset = builder.create_vector(&ip_addrs);
    let labels_offset = builder.create_vector(&labels);
    let platform_metadata_offset = builder.create_vector(&platform_metadata);

    let mut node = FlatNodeBuilder::new(&mut builder);

    node.add_name(name);
    node.add_namespace(namespace_);
    node.add_owner(owner);
    node.add_workload_name(workload_name);
    node.add_istio_version(istio_version);
    node.add_mesh_id(mesh_id);
    node.add_cluster_id(cluster_id);
    node.add_labels(labels_offset);
    node.add_platform_metadata(platform_metadata_offset);
    node.add_app_containers(app_containers_offset);
    node.add_instance_ips(ip_addrs_offset);

    let data = node.finish();
    builder.finish(data, None);
    let buf = builder.finished_data().to_vec();

    buf
}

fn deserialize_map(bytes: &[u8]) -> Vec<(String, String)> {
    let mut map = Vec::new();
    if bytes.is_empty() {
        return map;
    }
    let size = u32::from_le_bytes(<[u8; 4]>::try_from(&bytes[0..4]).unwrap()) as usize;
    let mut p = 4 + size * 8;
    for n in 0..size {
        let s = 4 + n * 8;
        let size = u32::from_le_bytes(<[u8; 4]>::try_from(&bytes[s..s + 4]).unwrap()) as usize;
        let key = bytes[p..p + size].to_vec();
        p += size + 1;
        let size = u32::from_le_bytes(<[u8; 4]>::try_from(&bytes[s + 4..s + 8]).unwrap()) as usize;
        let value = bytes[p..p + size].to_vec();
        p += size + 1;
        map.push((
            String::from_utf8(key).unwrap(),
            String::from_utf8(value).unwrap(),
        ));
    }
    map
}

fn serialize_to_string_deterministic(
    metadata: protobuf::well_known_types::any::Any,
    metadata_bytes: &mut Vec<u8>,
) {
    let mut mcs = CodedOutputStream::new(metadata_bytes);
    metadata.write_to(&mut mcs).unwrap();
}
