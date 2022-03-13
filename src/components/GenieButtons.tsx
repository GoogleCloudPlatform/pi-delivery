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

import { Grid } from "@mui/material";

const buttonCount = 8;

interface Props {
  active?: number;
}

export default function GenieButtons(props: Props = {}) {
  const { active } = props;

  return (
    <Grid
      container
      columns={buttonCount}
      className="genie-controls"
      spacing={0}
    >
      {[...Array(buttonCount)].map((_, i) => (
        <Grid item xs={1} key={i}>
          <button
            className={`color-${i}`}
            data-id={i}
            data-active={i === active ? "active" : ""}
          >
            <span>{i}</span>
          </button>
        </Grid>
      ))}
    </Grid>
  );
}
