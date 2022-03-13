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

import * as React from "react";
import { FormControl, InputLabel, Select, MenuItem } from "@mui/material";
import type { SelectChangeEvent } from "@mui/material";
import isEqual from "react-fast-compare";

type Props = {
  id: string;
  items: [string | number, string][];
  label?: string;
  value: string | number;
  onChange?:
    | ((event: SelectChangeEvent, child: React.ReactNode) => void)
    | undefined;
};

function Selector({ id, items, label, value, onChange }: Readonly<Props>) {
  const labelId = `${id}-label`;
  return (
    <FormControl fullWidth>
      <InputLabel id={labelId}>{label}</InputLabel>
      <Select
        labelId={labelId}
        label={label}
        id={id}
        value={String(value)}
        onChange={onChange}
      >
        {items.map((v) => (
          <MenuItem key={v[0]} value={v[0]}>
            {v[1]}
          </MenuItem>
        ))}
      </Select>
    </FormControl>
  );
}

export default React.memo(Selector, isEqual);
