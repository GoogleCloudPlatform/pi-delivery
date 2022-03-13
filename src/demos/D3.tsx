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

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import useD3Demo from "hooks/use-d3-demo";
import { DigitEventDetail, PiStream } from "lib/pi";
import DigitInput from "components/DigitInput";
import ControlButtons from "components/ControlButtons";
import { Grid } from "@mui/material";

export default function D3() {
  const ref = useRef<HTMLDivElement>(null);
  const { init, draw } = useD3Demo({
    ref: ref,
    width: ref.current?.clientWidth ?? 400,
  });
  const piStream = useMemo(() => new PiStream(undefined, { delayMs: 100 }), []);
  const [startDigit, setStartDigit] = useState(0);
  const [started, setStarted] = useState(false);
  const [seeking, setSeeking] = useState(false);

  // useResizeObserver(ref, (entry) => setWidth(entry.contentRect.width));

  const onDigit = useCallback(
    (e: CustomEvent<DigitEventDetail>) => {
      draw(e.detail.digit);
      if (!seeking) setStartDigit(e.detail.position);
    },
    [draw, seeking]
  );

  const start = useCallback(() => {
    piStream.seek(startDigit);
    piStream.start();
    setStarted(true);
  }, [piStream, startDigit]);

  const stop = useCallback(() => {
    piStream.stop();
    setStarted(false);
  }, [piStream]);

  const reset = useCallback(() => {
    piStream.stop();
    piStream.seek(0);
    init();
    setStartDigit(0);
    setStarted(false);
  }, [init, piStream]);

  const onChange = useCallback((v: number) => {
    setSeeking(true);
    setStartDigit(v);
  }, []);

  const onChangeCommitted = useCallback(
    (v: number) => {
      setSeeking(false);
      setStartDigit(v);
      piStream.seek(v);
    },
    [piStream]
  );

  useEffect(() => {
    piStream.addEventListener("digit", onDigit as EventListener);

    return () => {
      piStream.removeEventListener("digit", onDigit as EventListener);
    };
  }, [onDigit, piStream]);

  return (
    <>
      <Grid item xs={12}>
        <div ref={ref} style={{maxWidth: 800}}></div>
      </Grid>
      <Grid item xs={12}>
        <DigitInput
          id="d3-demo-start-digit"
          onChange={onChange}
          onChangeCommitted={onChangeCommitted}
          value={startDigit}
        />
      </Grid>
      <Grid item xs={12}>
        <ControlButtons
          onStart={start}
          onStop={stop}
          onReset={reset}
          started={started}
        />
      </Grid>
    </>
  );
}
