
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

const modpow = require("../modpow");

test("0^0 mod 2 = 1", () => {
  expect(modpow(0n, 0n, 2n)).toBe(1n);
});

test("4^13 mod 497 = 445", () => {
  expect(modpow(4n, 13n, 497n)).toBe(445n);
});

test("2^90 mod 13 = 12", () => {
  expect(modpow(2n, 90n, 13n)).toBe(12n);
});

test("12697823428734623421^238437482347238 mod 314 = 57", () => {
  expect(modpow(12697823428734623421n, 238437482347238n, 314n)).toBe(57n);
});
