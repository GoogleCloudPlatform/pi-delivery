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

import { useEffect, useState } from "react";

import type * as mm from "@magenta/music";

const checkpoint =
  "https://storage.googleapis.com/magentadata/js/checkpoints/piano_genie/model/epiano/stp_iq_auto_contour_dt_166006";
const temperature = 0.25;

export default function usePianoGenie(load = true): mm.PianoGenie | undefined {
  const [genie, setGenie] = useState<mm.PianoGenie>();

  useEffect(() => {
    if (!load) return;
    const promise = Promise.all([
      import("@magenta/music/esm/piano_genie"),
      import("@tensorflow/tfjs"),
    ])
      .then(([mmg, tf]) => {
        tf.disableDeprecationWarnings();
        const genie = new mmg.PianoGenie(checkpoint);
        return genie.initialize().then(() => genie);
      })
      .then((genie) => {
        genie.next(0, temperature);
        genie.resetState();
        setGenie(genie);
        return genie;
      });

    return () => {
      promise.then((genie) => {
        genie.dispose();
      });
    };
  }, [load]);

  return genie;
}
