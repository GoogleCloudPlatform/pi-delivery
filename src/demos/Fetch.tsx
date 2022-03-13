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

import React, { useState, useCallback, useContext, useEffect, useMemo } from "react";
import { Pi } from "lib/pi";
import ControlButtons from "components/ControlButtons";
import DigitInput from "components/DigitInput";
import {
  FormControl,
  FormControlLabel,
  FormLabel,
  Grid,
  Radio,
  RadioGroup,
} from "@mui/material";
import ResultField from "components/ResultField";
import PiContext from "contexts/PiContext";

const placeholderText = "Press 'Fetch'";

export default function Fetch() {
  const [text, setText] = useState(placeholderText);
  const [startDigit, setStartDigit] = useState(0);
  const { radix, setRadix, length } = useContext(PiContext);
  const pi = useMemo(() => new Pi(), []);

  const onFetch = useCallback(async () => {
    pi.radix = radix;
    setText(await pi.get(startDigit, 100));
  }, [pi, radix, startDigit]);

  const onRadixChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      setRadix(parseInt(e.target.value));
    },
    [setRadix]
  );

  useEffect(() => {
    setStartDigit((prev) => {
      if (prev > length) return length;
      return prev;
    });
  }, [length]);

  return (
    <>
      <Grid item xs={12}>
        <DigitInput
          id="fetch-starting-digit"
          value={startDigit}
          onChange={(n) => setStartDigit(n)}
        />
      </Grid>
      <Grid item xs={12}>
        <FormControl variant="outlined">
          <FormLabel id="fetch-base-label">Base</FormLabel>
          <RadioGroup
            row
            aria-labelledby="fetch-base-label"
            name="fetch-base-radio-group"
            value={radix}
            onChange={onRadixChange}
          >
            <FormControlLabel value="10" control={<Radio />} label="10" />
            <FormControlLabel value="16" control={<Radio />} label="16" />
          </RadioGroup>
        </FormControl>
      </Grid>
      <Grid item xs={12}>
        <ControlButtons startLabel="Fetch" onStart={onFetch} />
      </Grid>
      <Grid item xs={12}>
        <ResultField
          id="fetch-result"
          placeholder={"Click Fetch"}
          value={text}
        />
      </Grid>
    </>
  );
}
