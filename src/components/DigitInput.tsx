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

import React, { SyntheticEvent, useCallback, useContext } from "react";
import { FormLabel, Grid, Input, Slider } from "@mui/material";
import PiContext from "contexts/PiContext";
import isEqual from "react-fast-compare";

interface PropType {
  id: string;
  value: number;
  onChange?: (v: number) => void;
  onChangeCommitted?: (v: number) => void;
  disabled?: boolean;
}

function DigitInput({
  id,
  value,
  onChange,
  onChangeCommitted,
  disabled,
}: PropType) {
  const context = useContext(PiContext);

  const textFieldOnChange = useCallback(
    (
      e: React.FormEvent<HTMLInputElement | HTMLTextAreaElement> | undefined
    ) => {
      const v = parseInt(e?.currentTarget.value ?? "");
      if (isNaN(v)) return;
      if (onChange) {
        onChange(v);
      }
      if (onChangeCommitted) {
        onChangeCommitted(v);
      }
    },
    [onChange, onChangeCommitted]
  );

  const sliderOnChange = useCallback(
    (_e: Event, v: number | number[]) => {
      if (onChange) {
        onChange(v as number);
      }
    },
    [onChange]
  );

  const sliderOnChangeCommitted = useCallback(
    (_e: SyntheticEvent | Event, v: number | number[]) => {
      if (onChangeCommitted) {
        onChangeCommitted(v as number);
      }
    },
    [onChangeCommitted]
  );

  return (
    <Grid container spacing={3} alignItems="center">
      <Grid item xs="auto">
        <FormLabel htmlFor={id}>Starting Digit</FormLabel>
      </Grid>
      <Grid item xs>
        <Slider
          id={id}
          aria-labelledby={id}
          min={0}
          max={context.length}
          value={value}
          disabled={disabled}
          onChange={sliderOnChange}
          onChangeCommitted={sliderOnChangeCommitted}
        />
      </Grid>
      <Grid item xs="auto">
        <Input
          type="number"
          value={value}
          disabled={disabled}
          onChange={textFieldOnChange}
          inputProps={{
            min: 0,
            max: context.length,
            "aria-labelledby": id,
          }}
          sx={{ width: "17ch" }}
        />
      </Grid>
    </Grid>
  );
}

export default React.memo(DigitInput, isEqual);
