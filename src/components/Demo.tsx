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

import React, { Suspense, useMemo, useState } from "react";

import { Alert, Grid, LinearProgress } from "@mui/material";
import { useInView } from "react-intersection-observer";

import PiContext, * as pic from "contexts/PiContext";
import InViewContext from "contexts/InViewContext";
import { Pi } from "lib/pi";

interface Props {
  children: React.ReactNode;
}

export default function Demo({ children }: Props) {
  const { inView, ref } = useInView({ triggerOnce: true });
  const pi = useMemo(() => new Pi(), []);
  const [radix, setRadix] = useState(pic.defaults.radix);

  pi.radix = radix;
  const picValues = {
    length: pi.length,
    radix,
    setRadix,
  }

  return (
    <InViewContext.Provider value={inView}>
      <PiContext.Provider value={picValues}>
        <div ref={ref} className="demo">
          <Suspense
            fallback={
              <div>
                <Alert severity="info">Loading...</Alert>
                <LinearProgress />
              </div>
            }
          >
            <Grid container spacing={2}>
              {children}
            </Grid>
          </Suspense>
        </div>
      </PiContext.Provider>
    </InViewContext.Provider>
  );
}
