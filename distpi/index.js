// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

const functions = require('@google-cloud/functions-framework');
const calculatePiChunk = require("./bbp");

functions.http('httpCalc', (req, res) => {
    const offset = req.query.offset || req.body.offset || 0;
    const length = req.query.length || req.body.length || 100;

    const result = calculatePiChunk(offset, length);
    res.send(result);
});

exports.pubsubCalc = (evt, ctx) => {
    const message = Buffer.from(evt.data, "base64").toString();
    const params  = JSON.parse(message);

    const offset = params.start || 0;
    const length = params.size  || 100;

    const result = calculatePiChunk(offset, length);
}
