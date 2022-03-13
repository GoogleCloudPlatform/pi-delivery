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

import React, { useCallback, useState } from "react";
import { TextField, TextFieldProps } from "@mui/material";
import isEqual from "react-fast-compare";

export type Props = Omit<TextFieldProps, "onChange"> & {
  min?: number;
  max?: number;
  onChange?: (v: number) => void;
};

function BpmInput(props: Readonly<Props>) {
  const {
    type = "number",
    label = "BPM",
    variant = "outlined",
    onChange,
    min = 1,
    max,
    error,
    ...rest
  } = props;

  const [parseError, setParseError] = useState(false);

  const changed = useCallback(
    (e: React.ChangeEvent<HTMLTextAreaElement | HTMLInputElement>) => {
      if (e.target.value === "") {
        return;
      }

      let bpm = parseInt(e.target.value);
      if (isNaN(bpm)) {
        setParseError(true);
        return;
      }
      setParseError(false);
      if (max) bpm = Math.min(max, bpm);
      if (min) bpm = Math.max(min, bpm);
      if (onChange) onChange(bpm);
    },
    [max, min, onChange]
  );

  return (
    <TextField
      type={type}
      label={label}
      variant={variant}
      fullWidth
      onChange={changed}
      error={error || parseError}
      {...rest}
    />
  );
}

export default React.memo(BpmInput, isEqual);
