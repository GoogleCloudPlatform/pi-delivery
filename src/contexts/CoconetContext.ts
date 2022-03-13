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

import { scaleNames } from "lib/music";
import { createContext } from "react";

type setState<T> = React.Dispatch<React.SetStateAction<T>>;

export enum CoconetState {
  Initialized,
  LoadRequested,
  Loading,
  Loaded,
  HarmonizeRequested,
  Harmonizing,
  Harmonized,
}

export interface ICoconetContext {
  state: CoconetState;
  setState: setState<CoconetState>;
  startDigit: number;
  setStartDigit: setState<number>;
  instrument: number;
  scale: string;
  bpm: number;
  shouldPlay: boolean;
  loopPlay: boolean;
  useCoconet: boolean;
}

/* eslint-disable @typescript-eslint/no-empty-function */
export const defaults: ICoconetContext = {
  state: CoconetState.Initialized,
  setState: () => {},
  startDigit: 0,
  setStartDigit: () => {},
  instrument: 0,
  scale: scaleNames[0][0],
  bpm: 120,
  shouldPlay: false,
  loopPlay: false,
  useCoconet: true,
};

const CoconetContext = createContext(defaults);

export default CoconetContext;
