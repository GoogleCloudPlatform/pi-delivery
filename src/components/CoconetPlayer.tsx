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

import { useContext, useEffect, useState } from "react";

import type * as mm from "@magenta/music";
import noteMidi from "note-midi";

import MagentaVisualizer from "components/MagentaVisualizer";
import CoconetContext, { CoconetState } from "contexts/CoconetContext";
import useCoconetModel from "hooks/use-coconet-model";
import useMagentaPlayer, {
  State as PlayerState,
} from "hooks/use-magenta-player";
import * as Music from "lib/music";
import { Pi } from "lib/pi";
import { Box } from "@mui/material";
import InViewContext from "contexts/InViewContext";

const pi = new Pi();

const defaultBpm = 120;
const chunkLength = 32;
interface MelodyParameters {
  startDigit: number;
  instrument: number;
  bpm?: number;
  scale: string;
  length: number;
}

function createMelody({
  startDigit,
  instrument,
  bpm = defaultBpm,
  scale,
  length,
}: MelodyParameters): Promise<mm.INoteSequence> {
  const q = 4;
  return Promise.all([
    pi.get(startDigit % (pi.length - length + 1), length),
    import("@magenta/music/esm/protobuf"),
  ])
    .then(([digits, mmp]) => {
      const notes = new Array<mm.NoteSequence.Note>();
      for (let i = 0; i < digits.length; i++) {
        const d = parseInt(digits[i]);
        const scaleNote = Music.scales.get(scale)![d];
        const midiNote = noteMidi(scaleNote);
        notes.push(
          mmp.NoteSequence.Note.create({
            pitch: midiNote,
            quantizedStartStep: i * q,
            quantizedEndStep: (i + 1) * q,
            instrument: 0,
            program: instrument,
            velocity: 80,
          })
        );
      }
      const melody = mmp.NoteSequence.create({
        notes: notes,
        quantizationInfo: { stepsPerQuarter: 4 },
        totalQuantizedSteps: digits.length * q,
        tempos: [
          mmp.NoteSequence.Tempo.create({
            time: 0,
            qpm: bpm,
          }),
        ],
      });
      return melody;
    })
    .catch(() => {
      throw new Error("melody generation error (pi api unreachable?)");
    });
}

interface HarmonyParameters {
  model: mm.Coconet;
  melody: mm.INoteSequence;
}

function createHarmony({
  model,
  melody,
}: HarmonyParameters): Promise<mm.INoteSequence> {
  return Promise.all([
    import("@magenta/music/esm/core/sequences"),
    model.infill(melody),
  ]).then(([mms, sequence]) => {
    const harmony = mms.replaceInstruments(
      mms.mergeConsecutiveNotes(sequence),
      melody
    );
    return harmony;
  });
}

export default function CoconetPlayer() {
  const {
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
  } = useContext(CoconetContext);
  const inView = useContext(InViewContext);

  const [melody, setMelody] = useState<mm.INoteSequence>();
  const [nextMelody, setNextMelody] = useState<mm.INoteSequence>();
  const [harmony, setHarmony] = useState<mm.INoteSequence>();
  const [nextHarmony, setNextHarmony] = useState<mm.INoteSequence>();
  const [needsPrefetch, setNeedsPrefetch] = useState(false);

  const model = useCoconetModel(state !== CoconetState.Initialized);

  const {
    start,
    pause,
    stop,
    activeNote,
    state: playerState,
    atEnd: playerAtEnd,
  } = useMagentaPlayer({
    load: state !== CoconetState.Initialized,
    notes: useCoconet ? harmony : melody,
    loop: false,
    bpm,
  });

  // Start loading once visible.
  useEffect(() => {
    if (state === CoconetState.Initialized && inView)
      setState(CoconetState.LoadRequested);
  }, [inView, setState, state]);

  // Play if ready.
  useEffect(() => {
    if (shouldPlay && state === CoconetState.Harmonized && melody && harmony) {
      start();
    } else if (!shouldPlay && playerState === PlayerState.Started) {
      pause();
    }
  }, [harmony, melody, pause, playerState, shouldPlay, start, state]);

  // State machine...
  useEffect(() => {
    switch (state) {
      case CoconetState.LoadRequested:
        setState(CoconetState.Loading);
        stop();
        createMelody({
          startDigit,
          instrument,
          scale,
          length: chunkLength,
        }).then((m) => {
          setMelody(m);
          setState(CoconetState.Loaded);
        });
        break;
      case CoconetState.Loaded:
        setState(CoconetState.HarmonizeRequested);
        break;
      case CoconetState.HarmonizeRequested:
        if (!model || !melody) return;
        setState(CoconetState.Harmonizing);
        stop();
        createHarmony({ model, melody }).then((h) => {
          setHarmony(h);
          setState(CoconetState.Harmonized);
          setNeedsPrefetch(true);
        });
        break;
    }
  }, [instrument, melody, model, scale, setState, startDigit, state, stop]);

  // Prefetch next.
  useEffect(() => {
    if (state !== CoconetState.Harmonized || !model) return;
    if (needsPrefetch) {
      setNeedsPrefetch(false);
      createMelody({
        startDigit: startDigit + chunkLength,
        instrument,
        scale,
        length: chunkLength,
      })
        .then((m) => {
          setNextMelody(m);
          return createHarmony({ model, melody: m });
        })
        .then((h) => {
          setNextHarmony(h);
        });
    }
  }, [instrument, model, needsPrefetch, nextMelody, scale, startDigit, state]);

  // Move to the next section.
  useEffect(() => {
    if (playerAtEnd) {
      if (!loopPlay && nextMelody && nextHarmony) {
        setStartDigit((d) => (d + chunkLength) % (pi.length - chunkLength + 1));
        setMelody(nextMelody);
        setNextMelody(undefined);
        setHarmony(nextHarmony);
        setNextHarmony(undefined);
        setNeedsPrefetch(true);
      }
    }
  }, [loopPlay, nextHarmony, nextMelody, playerAtEnd, setStartDigit]);

  return (
    <Box
      height={270}
      minWidth={480}
      border={1}
      borderRadius={2}
      padding={1}
      
      bgcolor="#fff"
    >
      <MagentaVisualizer
        notes={useCoconet ? harmony : melody}
        height={270}
        width={480}
        activeNote={activeNote}
        noteColor="#90a4ae"
        activeNoteColor="#e53935"
      />
    </Box>
  );
}
