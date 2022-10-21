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

import Istio
import ArgumentParser

@main
struct CliIstio: ParsableCommand {
    @Option
    var heimdallURL = "https://af001efbdd27f4badba8b5e791d7417c-9dcfb5194711f057.elb.eu-central-1.amazonaws.com/config"
    
    @Option
    var clientID = "9dcfb5194711f057"
    
    @Option
    var clientSecret = "af001efbdd27f4badba8b5e791d7417c"
    
    mutating func run() throws {
        var error: NSError?
        
        let transport = IstioNewHTTPTransport(heimdallURL, clientID, clientSecret, &error)
        if let e = error {
            throw e
        }
        
        let response = try transport!.request("GET", url: "https://echo.demo:8080", body: "{}")
        
        print(String(data: response.body!, encoding: .utf8)!)
        
        transport!.close()
    }
}
