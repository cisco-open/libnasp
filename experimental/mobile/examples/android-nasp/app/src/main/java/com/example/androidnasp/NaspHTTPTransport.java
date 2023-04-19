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

package com.example.androidnasp;

import android.app.ProgressDialog;
import android.content.Context;
import android.os.AsyncTask;

import java.util.function.Function;

import nasp.HTTPTransport;
import nasp.Nasp;

public class NaspHTTPTransport {
    private static HTTPTransport httpTransport;

    public static HTTPTransport get() {
        return httpTransport;
    }

    public static void set(HTTPTransport httpTransport) {
        NaspHTTPTransport.httpTransport = httpTransport;
    }
}

class NaspHTTPTransportCreationTask extends AsyncTask<Void, Void, Exception> {
    private ProgressDialog dialog;
    private String heimdallURL;
    private String heimdallToken;
    private Runnable okAction;
    private Function<Exception, Void> errorAction;

    public NaspHTTPTransportCreationTask(Context context, String heimdallURL, String heimdallToken, Runnable okAction, Function<Exception, Void> errorAction) {
        dialog = new ProgressDialog(context);
        this.heimdallURL = heimdallURL;
        this.heimdallToken = heimdallToken;
        this.okAction = okAction;
        this.errorAction = errorAction;
    }

    @Override
    protected void onPreExecute() {
        dialog.setMessage("Connecting to Service Mesh...");
        dialog.show();
    }

    @Override
    protected Exception doInBackground(Void... args) {
        String authorizationToken = "TODO";

        try {
            HTTPTransport httpTransport = Nasp.newHTTPTransport(heimdallURL, heimdallToken);
            NaspHTTPTransport.set(httpTransport);
        } catch (Exception e) {
            return e;
        }

        return null;
    }

    @Override
    protected void onPostExecute(Exception error) {
        // do UI work here
        if (dialog.isShowing()) {
            dialog.dismiss();
        }

        if (error != null) {
            errorAction.apply(error);
        } else {
            okAction.run();
        }
    }
}