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

import ctypes, sys
import getopt
import io
import threading
from concurrent.futures import ThreadPoolExecutor

import requests
import urllib3


class GoError(ctypes.Structure):
    _fields_ = [("error_msg", ctypes.c_char_p)]

    @property
    def msg(self):
        if self.error_msg:
            return ctypes.string_at(self.error_msg).decode("utf-8")

        return None


class NewHTTPTransportReturn(ctypes.Structure):
    _fields_ = [("r0", ctypes.c_ulonglong), ("r1", ctypes.POINTER(GoError))]

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        if self.r1:
            nasp.free_go_error(self.r1)

    # id of the created HTTP transport
    @property
    def id(self):
        return self.r0

    # pointer to error message if an error occurred
    @property
    def error(self):
        return self.r1


class GoHttpHeader(ctypes.Structure):
    _fields_ = [("key", ctypes.c_char_p), ("value", ctypes.c_char_p)]

    @property
    def name(self):
        if self.key:
            return ctypes.string_at(self.key).decode("utf-8")

        return None

    @property
    def val(self):
        if self.value:
            return ctypes.string_at(self.value).decode("utf-8")

        return None


class GoHttpHeaders(ctypes.Structure):
    _fields_ = [("items", ctypes.POINTER(GoHttpHeader)), ("len", ctypes.c_uint)]


class GoHttpResponse(ctypes.Structure):
    _fields_ = [("status_code", ctypes.c_int), ("version", ctypes.c_char_p), ("headers", GoHttpHeaders),
                ("body", ctypes.c_char_p)]

    @property
    def status(self):
        return int(self.status_code)

    @property
    def http_version(self):
        if self.version:
            return ctypes.string_at(self.version)

        return None

    @property
    def headers_dict(self):
        if self.headers:
            d = {}
            for i in range(self.headers.len):
                d[self.headers.items[i].name] = self.headers.items[i].val

            return d

        return {}

    @property
    def response_body(self):
        if self.body:
            return ctypes.string_at(self.body)

        return None


class SendHTTPRequestReturn(ctypes.Structure):
    _fields_ = [("r0", ctypes.POINTER(GoHttpResponse)), ("r1", ctypes.POINTER(GoError))]

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        if self.r0:
            nasp.free_go_http_response(self.r0)

        if self.r1:
            nasp.free_go_error(self.r1)

    @property
    def response(self):
        return self.r0

    @property
    def error(self):
        return self.r1


nasp = ctypes.cdll.LoadLibrary("./nasp.so")

nasp.free_go_error.argtypes = [ctypes.POINTER(GoError)]

nasp.NewHTTPTransport.argtypes = [ctypes.c_char_p, ctypes.c_char_p, ctypes.c_char_p]
nasp.NewHTTPTransport.restype = NewHTTPTransportReturn

nasp.SendHTTPRequest.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p, ctypes.c_char_p, ctypes.c_char_p]
nasp.SendHTTPRequest.restype = SendHTTPRequestReturn

nasp.free_go_http_response.argtypes = [ctypes.POINTER(GoHttpResponse)]

nasp.CloseHTTPTransport.argtypes = [ctypes.c_ulonglong]


def new_http_transport(heimdall_url: str, client_id: str, client_secret: str):
    return nasp.NewHTTPTransport(heimdall_url.encode("utf-8"), client_id.encode("utf-8"),
                                 client_secret.encode("utf-8"))


def send_http_request(http_transport_id: ctypes.c_ulonglong,
                      method: str, url: str, body: bytes) -> SendHTTPRequestReturn:
    return nasp.SendHTTPRequest(
        http_transport_id,
        method.encode("utf-8"),
        url.encode("utf-8"),
        body,
    )


def close_http_transport(http_transport_id: ctypes.c_ulonglong):
    nasp.CloseHTTPTransport(http_transport_id)


class NaspHTTPTransportAdapter(requests.adapters.BaseAdapter):
    def __init__(self, http_transport_id):
        super().__init__()

        self.http_transport_id = http_transport_id

    def send(self, request, stream=False, timeout=None, verify=True, cert=None, proxies=None):
        response = requests.Response()

        with send_http_request(self.http_transport_id, request.method, request.url, request.body) as ret:
            raw = urllib3.HTTPResponse(
                body=io.BytesIO(ret.response.contents.response_body) if ret.response else "",
                headers=ret.response.contents.headers_dict if ret.response else {},
                request_method=request.method,
                request_url=request.url,
                preload_content=False,
                status=ret.response.contents.status if ret.response else 0,
            )

            if ret.error:
                raw.reason = ret.error.contents.msg

            response.raw = raw
            response.url = request.url
            response.headers = requests.structures.CaseInsensitiveDict(raw.headers)
            response.encoding = requests.utils.get_encoding_from_headers(response.headers or {})
            response.request = request
            response.connection = self
            response.status_code = raw.status
            response.reason = raw.reason

        return response

    def close(self):
        close_http_transport(self.http_transport_id)


def main(argv):
    request_url = "http://catalog.demo.svc.cluster.local:8080"
    request_count = 5
    heimdall_url = "https://52.59.158.146:443/config"
    try:
        opts, args = getopt.getopt(argv, "h:", ["request-url=", "request-count=", "heimdall-url="])
    except getopt.GetoptError:
        print('main.py --request-url <request-url> --request-count <request-count> --heimdall-url <heimdall-url>')
        sys.exit(2)

    for opt, arg in opts:
        if opt == "-h":
            sys.exit()
        elif opt == "--request-url":
            request_url = arg
        elif opt == "--request-count":
            request_count = int(arg)
        elif opt == "--heimdall-url":
            heimdall_url = arg

    client_id = "test-http-16362813-F46B-41AC-B191-A390DB1F6BDF"
    client_secret = "16362813-F46B-41AC-B191-A390DB1F6BDF"

    with new_http_transport(heimdall_url, client_id, client_secret) as result:
        if result.error:
            err_msg = result.error.contents.msg
            sys.exit(err_msg)

    with requests.Session() as s:
        s.mount("http://", NaspHTTPTransportAdapter(result.id))

        with ThreadPoolExecutor() as executor:
            def http_get():
                with s.get(request_url) as resp:
                    print("thread[", threading.get_native_id(), "]\t", resp.status_code, "\t", resp.text)

            for i in range(request_count):
                executor.submit(http_get)


if __name__ == "__main__":
    main(sys.argv[1:])
