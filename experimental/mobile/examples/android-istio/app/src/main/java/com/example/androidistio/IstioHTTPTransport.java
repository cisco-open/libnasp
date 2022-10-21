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

package com.example.androidistio;

import android.app.ProgressDialog;
import android.content.Context;
import android.os.AsyncTask;
import android.provider.Settings;

import java.util.function.Function;

import istio.HTTPTransport;
import istio.Istio;

public class IstioHTTPTransport {
    private static HTTPTransport httpTransport;

    public static HTTPTransport get() {
        return httpTransport;
    }

    public static void set(HTTPTransport httpTransport) {
        IstioHTTPTransport.httpTransport = httpTransport;
    }
}

class IstioHTTPTransportCreationTask extends AsyncTask<Void, Void, Exception> {
    private ProgressDialog dialog;
    private String heimdallURL;
    private Runnable okAction;
    private Function<Exception, Void> errorAction;

    public IstioHTTPTransportCreationTask(Context context, String heimdallURL, Runnable okAction, Function<Exception, Void> errorAction) {
        dialog = new ProgressDialog(context);
        this.heimdallURL = heimdallURL;
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
        String clientID = Settings.Secure.getString(dialog.getContext().getContentResolver(), Settings.Secure.ANDROID_ID);
        String clientSecret = clientID;

        try {
            HTTPTransport httpTransport = Istio.newHTTPTransport(heimdallURL, clientID, clientSecret);
            IstioHTTPTransport.set(httpTransport);
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