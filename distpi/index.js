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

const modpow = require("./modpow");

exports.httpCalc = (req, res) => {
    let offset = req.query.offset || 0;
    let length = req.query.length || 100;

    let result = calculatePiChunk(offset, length);
    res.send(result);
}

exports.pubsubCalc = (evt, ctx) => {
    const message = Buffer.from(evt.data, "base64").toString();
    const params  = JSON.parse(message);
    const result = calculatePiChunk(params.start, params.size);
}

function summation(j, n, d, mask) {
    const shift = d << 2n;

    let left = 0n;

    for(let k = 0n; k <= n; k++) {
        let r = k * 8n + j;
        let exponent = n - k;
        left = (left + (modpow(16n, n - k, r) << shift) / r) & mask;
    }

    let right = 0n;

    for(let k = n + 1n; ; k++) {
        let rnew = right + 16n ** (d + n - k) / (k * 8n + j);
        if(right === rnew) { break; }
        right = rnew;
    }

    return left + right;
}

function calculatePiChunk(offset, length) {
    console.log("calculating " + length + " digits starting at offset " + offset);

    const biOffset = BigInt(offset) - 1n;
    const biLength = BigInt(length);
    const mask = 16n ** biLength - 1n;

    const t1 = summation(1n, biOffset, biLength, mask);
    const t2 = summation(4n, biOffset, biLength, mask);
    const t3 = summation(5n, biOffset, biLength, mask);
    const t4 = summation(6n, biOffset, biLength, mask);

    const total = (t1 * 4n) - (t2 * 2n) - t3 - t4;

    const result = (total & mask).toString(16);
    console.log("result: " + result);
    return result;
}
