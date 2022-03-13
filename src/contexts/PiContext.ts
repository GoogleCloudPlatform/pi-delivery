/**
 * Copyright 2022 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { Pi } from "lib/pi";
import { createContext } from "react";

type setState<T> = React.Dispatch<React.SetStateAction<T>>;
export interface IPiContext {
  length: number;
  radix: number;
  setRadix: setState<number>;
}

const pi = new Pi();

/* eslint-disable @typescript-eslint/no-empty-function */
export const defaults: IPiContext = {
  length: pi.length,
  radix: 10,
  setRadix: () => {},
};

const PiContext = createContext(defaults);

export default PiContext;
