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

export default function useCoconetModel(
  load: boolean,
  url = "https://storage.googleapis.com/magentadata/js/checkpoints/coconet/bach"
): mm.Coconet | undefined {
  const [model, setModel] = useState<mm.Coconet>();

  useEffect(() => {
    if (!load) return;

    const promise = Promise.all([
      import("@tensorflow/tfjs-backend-webgl"),
      import("@tensorflow/tfjs-backend-cpu"),
      import("@magenta/music/esm/coconet"),
    ])
      .then(([, , mmc]) => {
        const m = new mmc.Coconet(url);
        return m.initialize().then(() => m);
      })
      .then((m) => {
        setModel(m);
        return m;
      });

    return () => {
      promise.then((m) => {
        m.dispose();
      });
    };
  }, [url, load]);

  return model;
}
