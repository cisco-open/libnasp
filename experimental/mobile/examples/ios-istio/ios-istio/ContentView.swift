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

import SwiftUI
import Istio

enum Method: String, CaseIterable, Identifiable {
    case GET, POST, DELETE
    var id: Self { self }
}

let defaultConnectText = "Press Connect for Istio HTTP request"

struct ContentView: View {
    @State var configURL = "https://localhost:16443/config"
    @State var url = "https://echo.testing"
    @State var method = Method.GET
    @State var requestBody = ""
    @State var response = defaultConnectText
    @State var transport: IstioHTTPTransport?
    @State var selection : String?
    @State var showError = false
    @State var errorMessage = ""
    @State var bombard = false

    func newTransport() throws -> IstioHTTPTransport {
        var error: NSError?
        let clientID = UIDevice.current.identifierForVendor?.uuidString
        let clientSecret = UIDevice.current.identifierForVendor?.uuidString

        let transport = IstioNewHTTPTransport(configURL, clientID, clientSecret, &error)
        if let e = error {
            throw e
        }

        return transport!
    }
    
    func request() async throws -> String {
        let response = try transport!.request(method.rawValue, url: url, body: requestBody)
        return "HTTP \(response.statusCode)\n" + String(decoding: response.body!, as: UTF8.self)
    }
    
    var body: some View {
        let startForm = Form {
            TextField("Heimdall URL",
                      text: $configURL,
                      prompt: Text("Heimdall URL")).padding()
            Button("Connect", action: {
                selection = "progress"
            })
        }.onAppear {
            self.transport?.close()
        }.onDisappear {
            do {
                self.transport = try newTransport()
                selection = "form"
            } catch {
                showError = true
                errorMessage = "\(error)"
            }
        }
        
        let progress = ProgressView("Connecting to service mesh")
        
        let form = VStack {
            Form {
                TextField("URL", text: $url, prompt: Text("URL"))
                
                Picker("Method", selection: $method) {
                    ForEach(Method.allCases) { method in
                        Text(method.rawValue).tag(method)
                    }
                }
                
                TextField("Body", text: $requestBody, prompt: Text("Request Body"))
                
                Button("Send", action: {
                    response = "Connecting..."
                    Task {
                        do {
                            response = try await request()
                        } catch {
                            response = "Unexpected error: \(error)."
                        }
                    }
                })
                
                Button("Bombard", action: {
                    bombard = true
                    Timer.scheduledTimer(withTimeInterval: 2, repeats: true) { timer in
                        response = "Connecting..."
                        Task {
                            do {
                                response = try await request()
                                if !bombard {
                                    timer.invalidate()
                                }
                            } catch {
                                response = "Unexpected error: \(error)."
                            }
                        }
                    }
                }).disabled(bombard)

                Button("Stop", action: {
                    bombard = false
                }).disabled(bombard == false)
            }
            
            TextEditor(text: $response)
                .foregroundColor(Color.gray)
                .font(.custom("HelveticaNeue", size: 10))
                .textFieldStyle(.roundedBorder)
                .padding()
                .onAppear {
                    response = defaultConnectText
                }
        }
        
        NavigationStack() {
            startForm
            NavigationLink("", destination: progress, tag: "progress", selection: $selection)
            NavigationLink("", destination: form, tag: "form", selection: $selection)
        }.alert(isPresented: $showError) {
            Alert(title: Text("Failed to create NASP transport"), message: Text(errorMessage), dismissButton: .default(Text("Got it!")))
        }
    }
}
