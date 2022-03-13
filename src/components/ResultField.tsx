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

import React, { useCallback, useEffect, useState } from "react";
import {
  FormControl,
  IconButton,
  InputAdornment,
  InputLabel,
  OutlinedInput,
  OutlinedInputProps,
  Tooltip,
} from "@mui/material";
import ContentCopyIcon from "@mui/icons-material/ContentCopy";

export default function ResultField({
  id,
  rows = 3,
  value,
  placeholder,
  label,
}: OutlinedInputProps) {
  const copyToClipboard = useCallback(() => {
    navigator.clipboard.writeText(String(value)).then(() => {
      setTooltipOpen(true);
    });
  }, [value]);
  const [tooltipOpen, setTooltipOpen] = useState(false);

  const copyToClipboardMouseDown = useCallback((e: React.MouseEvent) => {
    e.preventDefault();
  }, []);

  const handleTooltipClose = useCallback(() => {
    setTooltipOpen(false);
  }, []);

  useEffect(() => {
    if (!tooltipOpen) return;
    const timer = setTimeout(() => setTooltipOpen(false), 3000);

    return () => {
      clearTimeout(timer);
    };
  }, [tooltipOpen]);

  return (
    <FormControl variant="outlined" fullWidth>
      <InputLabel htmlFor={id}>{label}</InputLabel>
      <OutlinedInput
        id={id}
        rows={rows}
        readOnly={true}
        multiline
        sx={{ fontFamily: "monospace" }}
        value={value}
        placeholder={placeholder}
        label={label}
        endAdornment={
          <InputAdornment position="end">
            <Tooltip
              title="Copied!"
              disableFocusListener
              disableHoverListener
              disableTouchListener
              open={tooltipOpen}
              onClose={handleTooltipClose}
            >
              <IconButton
                aria-label="copy to clipboard"
                onClick={copyToClipboard}
                onMouseDown={copyToClipboardMouseDown}
                edge="end"
              >
                <ContentCopyIcon />
              </IconButton>
            </Tooltip>
          </InputAdornment>
        }
      />
    </FormControl>
  );
}
