
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

const BBP = require("../bbp");

test("no args -> 100 digits @ position 0", () => {
    expect(BBP()).toBe("3243f6a8885a308d313198a2e03707344a4093822299f31d0082efa98ec4e6c89452821e638d01377be5466cf34e90c6cc0a");
});

test("50 digits @ position 0", () => {
  expect(BBP(0, 50)).toBe("3243f6a8885a308d313198a2e03707344a4093822299f31d02");
});

test("50 digits @ position 100,000", () => {
    expect(BBP(100000, 50)).toBe("535ea16c406363a30bf0b2e693992b58f7205a7232c4168840");
});
