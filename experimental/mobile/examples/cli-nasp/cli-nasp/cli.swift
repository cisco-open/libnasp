// Copyright (c) 2022 Cisco and/or its affiliates. All rights reserved.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       https://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

import Nasp
import ArgumentParser

@main
struct CliNasp: ParsableCommand {
    @Option
    var heimdallURL = "https://a2ee965a005324185b398968d6cc7fae-859cb095d9ee0179.elb.eu-central-1.amazonaws.com/config"
    
    @Option
    var heimdallToken = ""
    
    mutating func run() throws {
        var error: NSError?
        
        let transport = NaspNewHTTPTransport(heimdallURL, heimdallToken, &error)
        if let e = error {
            throw e
        }
        
        let response = try transport!.request("GET", url: "https://echo.demo:8080", body: "{}")
        
        print(String(data: response.body!, encoding: .utf8)!)
        
        transport!.close()
    }
}
