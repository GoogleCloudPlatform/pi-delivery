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

import { useCallback, useContext, useEffect, useMemo, useState } from "react";

import { Fade, Grid, LinearProgress } from "@mui/material";
import ControlButtons from "components/ControlButtons";
import DigitInput from "components/DigitInput";
import BpmInput from "components/BpmInput";
import InViewContext from "contexts/InViewContext";
import { DigitEventDetail, Pi, PiStream } from "lib/pi";
import GeniePiano from "components/GeniePiano";
import { bpmToDelayMs } from "lib/music";

const defaultBpm = 314;

const pi = new Pi();

export default function Genie() {
  const [playing, setPlaying] = useState(false);
  const [startDigit, setStartDigit] = useState(0);
  const [seeking, setSeeking] = useState(false);
  const [bpm, setBpm] = useState(defaultBpm);
  const [ready, setReady] = useState(false);
  const [activeButton, setActiveButton] = useState<number>();
  const inView = useContext(InViewContext);
  const piStream = useMemo(() => new PiStream(), []);

  const start = useCallback(() => {
    setPlaying(true);
  }, []);

  const stop = useCallback(() => {
    setActiveButton(undefined);
    setPlaying(false);
  }, []);

  const onChange = useCallback((n) => {
    setStartDigit(n);
    setSeeking(true);
  }, []);

  const onChangeCommitted = useCallback(
    (n) => {
      setStartDigit(n);
      piStream.seek(n);
      setSeeking(false);
    },
    [piStream]
  );

  const onBpmChange = useCallback((v: number) => {
    setBpm(v);
  }, []);

  const onDigit = useCallback(
    (e: CustomEvent<DigitEventDetail>) => {
      setActiveButton((prev) => {
        if (e.detail.digit > 7)
          return prev ?? 3;
        return e.detail.digit;
      });
      if (!seeking) setStartDigit(e.detail.position);

      if (e.detail.position >= pi.length) {
        piStream.seek(0);
        piStream.start();
      }
    },
    [piStream, seeking]
  );

  useEffect(() => {
    if (playing) {
      piStream.start();
    } else {
      piStream.stop();
    }
  }, [piStream, playing]);

  useEffect(() => {
    piStream.delayMs = bpmToDelayMs(bpm);
  }, [bpm, piStream]);

  useEffect(() => {
    piStream.addEventListener("digit", onDigit as EventListener);
    return () => {
      piStream.removeEventListener("digit", onDigit as EventListener);
    };
  }, [onDigit, piStream]);

  const onReady = useCallback(() => {
    setReady(true);
  }, []);

  return (
    <>
      <Grid item xs={12}>
        <GeniePiano
          activeButton={activeButton}
          onReady={onReady}
          inView={inView}
        ></GeniePiano>
      </Grid>
      {useMemo(
        () => (
          <Grid item xs={12}>
            <Fade in={!ready}>
              <LinearProgress />
            </Fade>
          </Grid>
        ),
        [ready]
      )}
      <Grid item xs={12}>
        <DigitInput
          id="genie-piano-starting-digit"
          onChange={onChange}
          onChangeCommitted={onChangeCommitted}
          value={startDigit}
        />
      </Grid>
      <Grid item xs={6}>
        <BpmInput id="genie-piano-bpm" value={bpm} onChange={onBpmChange} />
      </Grid>
      <Grid item xs={12}>
        <ControlButtons
          onStart={start}
          onStop={stop}
          started={playing}
          disabled={!ready}
        />
      </Grid>
    </>
  );
}
