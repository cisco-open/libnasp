#!/usr/bin/env python3
# encoding: utf-8


# Copyright (c) 2022 Cisco and/or its affiliates. All rights reserved.
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#       https://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.


import getopt
import sys
import threading
import time
from concurrent.futures import *

import requests

from nasp import nasp


def main(argv):
    request_url = "http://echo.testing:80"
    request_count = 5
    heimdall_url = "https://localhost:16443/config"
    use_push_gateway = None
    push_gateway_address = "push-gw-prometheus-pushgateway.prometheus-pushgateway.svc.cluster.local:9091"
    client_sleep_seconds = 0

    try:
        opts, args = getopt.getopt(argv, "h:", ["request-url=", "request-count=", "heimdall-url=",
                                                "use-push-gateway", "push-gateway-address=", "client-sleep="])
    except getopt.GetoptError:
        print('main.py --request-url <request-url> --request-count <request-count> --heimdall-url <heimdall-url>'
              '--use-push-gateway --push-gateway-address <push-gateway-address> --client-sleep <seconds>')
        sys.exit(2)

    for opt, arg in opts:
        if opt == "-h":
            sys.exit()
        elif opt == "--request-url":
            request_url = arg
        elif opt == "--request-count":
            request_count = int(arg)
        elif opt == "--client-sleep":
            client_sleep_seconds = int(arg)
        elif opt == "--heimdall-url":
            heimdall_url = arg
        elif opt == "--push-gateway-address":
            push_gateway_address = arg
        elif opt == "--use-push-gateway":
            use_push_gateway = arg

    client_id = "test-http-16362813-F46B-41AC-B191-A390DB1F6BDF"
    client_secret = "16362813-F46B-41AC-B191-A390DB1F6BDF"

    with nasp.new_http_transport(heimdall_url, client_id, client_secret,
                            use_push_gateway is not None, push_gateway_address) as http_transport:
        if http_transport.error:
            err_msg = http_transport.error.contents.msg
            sys.exit(err_msg)

        with requests.Session() as s:
            s.mount("http://", nasp.NaspHTTPTransportAdapter(http_transport.id))

            with ThreadPoolExecutor() as executor:
                def http_get():
                    with s.get(request_url) as resp:
                        print("thread[", threading.get_native_id(), "]\t", resp.status_code, "\t", resp.text)

                for i in range(request_count):
                    executor.submit(http_get)

        time.sleep(client_sleep_seconds)


if __name__ == "__main__":
    main(sys.argv[1:])
