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

import { Button, Stack, styled } from "@mui/material";
import isEqual from "react-fast-compare";

interface PropType {
  onStart: () => void;
  onStop: () => void;
  onReset: () => void;
  started: boolean;
  disabled: boolean;
  startLabel: string;
  stopLabel: string;
  resetLabel: string;
}

const ControlButton = styled(Button)(() => ({
  width: 120,
  backfaceVisibility: "hidden",
}));

function ControlButtons({
  onStart,
  onStop,
  onReset,
  started = false,
  disabled,
  startLabel = "Start",
  stopLabel = "Stop",
  resetLabel = "Reset",
}: Partial<PropType>) {
  return (
    <Stack direction="row" spacing={1}>
      <ControlButton
        variant="contained"
        onClick={onStart}
        disabled={started || disabled}
      >
        {startLabel}
      </ControlButton>

      {onStop ? (
        <ControlButton
          variant="outlined"
          onClick={onStop}
          disabled={!started || disabled}
        >
          {stopLabel}
        </ControlButton>
      ) : null}

      {onReset ? (
        <ControlButton variant="outlined" onClick={onReset} disabled={disabled}>
          {resetLabel}
        </ControlButton>
      ) : null}
    </Stack>
  );
}

export default React.memo(ControlButtons, isEqual);
