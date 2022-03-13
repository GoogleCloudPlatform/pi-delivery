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

import { useState, useCallback } from "react";

import {
  Fade,
  FormControlLabel,
  FormGroup,
  Grid,
  LinearProgress,
  Switch,
} from "@mui/material";

import Selector from "components/Selector";
import DigitInput from "components/DigitInput";
import ControlButtons from "components/ControlButtons";
import CoconetContext, {
  CoconetState,
  ICoconetContext,
} from "contexts/CoconetContext";
import CoconetPlayer from "components/CoconetPlayer";
import * as Music from "lib/music";
import BpmInput from "components/BpmInput";

const scaleNames = Music.scaleNames.filter((v) =>
  ["c-major", "e-major", "am-pentatonic"].includes(v[0])
);

const defaultBpm = 120;
const minBpm = 1,
  maxBpm = 200;

export default function Coconet() {
  const [state, setState] = useState(CoconetState.Initialized);
  const [instrument, setInstrument] = useState(0);
  const [scale, setScale] = useState(scaleNames[0][0]);
  const [startDigit, setStartDigit] = useState(0);
  const [bpm, setBpm] = useState(defaultBpm);
  const [useCoconet, setUseCoconet] = useState(true);
  const [shouldPlay, setShouldPlay] = useState(false);
  const [loopPlay, setLoopPlay] = useState(false);

  const contextValue: ICoconetContext = {
    state,
    setState,
    startDigit,
    setStartDigit,
    instrument,
    scale,
    bpm,
    shouldPlay,
    loopPlay,
    useCoconet,
  };

  const reset = useCallback(() => {
    setShouldPlay(false);
    setStartDigit(0);
    setInstrument(0);
    setScale(scaleNames[0][0]);
    setState(CoconetState.LoadRequested);
  }, []);

  const onScaleChange = useCallback((e) => {
    setScale(e.target.value);
    setState(CoconetState.LoadRequested);
  }, []);

  const onInstrumentChange = useCallback((e) => {
    setInstrument(e.target.value);
    setState(CoconetState.LoadRequested);
  }, []);

  const onBpmChange = useCallback((v: number) => {
    setBpm(v);
  }, []);

  return (
    <>
      <CoconetContext.Provider value={contextValue}>
        <Grid item xs={12} justifyContent="center">
          <CoconetPlayer />
        </Grid>
        <Grid item xs={12}>
          <Fade
            in={
              state === CoconetState.Harmonizing ||
              state === CoconetState.HarmonizeRequested
            }
          >
            <LinearProgress />
          </Fade>
        </Grid>
        <Grid item xs={12}>
          <DigitInput
            id="coconet-starting-digit"
            value={startDigit}
            onChange={(v) => {
              setStartDigit(v);
            }}
            onChangeCommitted={(v) => {
              setStartDigit(v);
              setState(CoconetState.LoadRequested);
            }}
          />
        </Grid>
        <Grid item xs={6} lg={4}>
          <BpmInput
            value={bpm}
            onChange={onBpmChange}
            fullWidth
            min={minBpm}
            max={maxBpm}
          />
        </Grid>
        <Grid item xs={6} lg={4}>
          <Selector
            id="coconet-scale-selector"
            label="Scale"
            onChange={onScaleChange}
            value={scale}
            items={scaleNames}
          ></Selector>
        </Grid>
        <Grid item xs={6} lg={4}>
          <Selector
            id="coconet-instrument-selector"
            label="Instrument"
            onChange={onInstrumentChange}
            value={instrument}
            items={Music.instruments}
          />
        </Grid>
        <Grid item xs={12}>
          <FormGroup>
            <FormControlLabel
              control={
                <Switch
                  checked={useCoconet}
                  onChange={(e) => {
                    setUseCoconet(e.target.checked);
                  }}
                />
              }
              label="Use machine learning"
            />
            <FormControlLabel
              control={
                <Switch
                  checked={loopPlay}
                  onChange={(e) => {
                    setLoopPlay(e.target.checked);
                  }}
                />
              }
              label="Loop play"
            />
          </FormGroup>
        </Grid>

        <Grid item xs={12}>
          <ControlButtons
            started={shouldPlay}
            disabled={
              state === CoconetState.Harmonizing ||
              state === CoconetState.HarmonizeRequested ||
              state === CoconetState.LoadRequested ||
              state === CoconetState.Loading
            }
            onStart={() => {
              if (state === CoconetState.Initialized)
                setState(CoconetState.LoadRequested);
              setShouldPlay(true);
            }}
            onStop={() => {
              setShouldPlay(false);
            }}
            onReset={reset}
          />
        </Grid>
      </CoconetContext.Provider>
    </>
  );
}
