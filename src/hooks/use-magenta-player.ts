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

import { useState, useEffect, useCallback } from "react";
import type * as mm from "@magenta/music";
import * as Tone from "tone";
import { highestMidiNote, lowestMidiNote } from "lib/music";

interface Props {
  load: boolean;
  url: string;
  notes: mm.INoteSequence;
  loop: boolean;
  bpm: number;
}

export enum State {
  Stopped,
  Started,
  Paused,
}

interface Player {
  start: () => void;
  pause: () => void;
  stop: () => void;
  playNoteUp: (note: mm.NoteSequence.INote) => void;
  playNoteDown: (note: mm.NoteSequence.INote) => void;
  atEnd: boolean;
  state: State;
  activeNote: mm.NoteSequence.INote | undefined;
  ready: boolean;
}

const defaultBpm = 120;

export default function useMagentaPlayer({
  load = true,
  url = "https://storage.googleapis.com/magentadata/js/soundfonts/sgm_plus",
  notes,
  loop = true,
  bpm = defaultBpm,
}: Partial<Props> = {}): Player {
  const [player, setPlayer] = useState<mm.SoundFontPlayer>();
  const [state, setState] = useState(State.Stopped);
  const [atEnd, setAtEnd] = useState(false);
  const [currentPlayTime, setCurrentPlayTime] = useState(0);
  const [activeNote, setActiveNote] = useState<mm.NoteSequence.INote>();

  const playNotes = useCallback(() => {
    if (!player || !notes) return;

    setState(State.Started);
    setAtEnd(false);
    player.start(notes, 120, currentPlayTime).then(() => {
      setAtEnd(true);
      setCurrentPlayTime(0);
      setState(State.Stopped);
    });
  }, [currentPlayTime, notes, player]);

  const start = useCallback(() => {
    if (!player) return;

    if (state === State.Paused) {
      player.resume();
      setState(State.Started);
    } else if (state === State.Stopped) {
      playNotes();
    }
  }, [playNotes, player, state]);

  const pause = useCallback(() => {
    if (state === State.Started) {
      player?.pause();
      setState(State.Paused);
    }
  }, [player, state]);

  const stop = useCallback(() => {
    setCurrentPlayTime(0);
    player?.stop();
    setState(State.Stopped);
  }, [player]);

  const playNoteUp = useCallback(
    (note: mm.NoteSequence.INote) => {
      player?.playNoteUp(note);
    },
    [player]
  );

  const playNoteDown = useCallback(
    (note: mm.NoteSequence.INote) => {
      player?.playNoteDown(note);
    },
    [player]
  );

  useEffect(() => {
    if (!load) return;
    const promise = import("@magenta/music/esm/core/player").then(
      async (mmp) => {
        const player = new mmp.SoundFontPlayer(
          url,
          undefined,
          undefined,
          undefined,
          {
            run: (n) => {
              if (n.startTime) setCurrentPlayTime(n.startTime);
              setActiveNote(n);
            },
            stop: () => {
              if (!player) return;
            },
          }
        );
        const notes: mm.NoteSequence.INote[] = [];
        for (let i = lowestMidiNote; i <= highestMidiNote; i++) {
          notes.push({ pitch: i });
        }
        await player.loadSamples({ notes });
        setPlayer(player);
        return player;
      }
    );
    return () => {
      promise.then((p) => {
        p.stop();
        setState(State.Stopped);
      });
    };
  }, [load, url]);

  useEffect(() => {
    if (atEnd && state === State.Started) {
      if (loop) {
        playNotes();
      } else {
        setState(State.Stopped);
      }
    }
  }, [loop, playNotes, atEnd, state]);

  useEffect(() => {
    // The tempo control of the player is kinda broken so let's tweak the
    // underlying Tone.js...
    // We do this at every rendering because it's pretty lightweight and
    // changing it right after player.start() doesn't work.
    if (bpm) Tone.Transport.bpm.value = bpm;
  });

  useEffect(() => {
    player?.stop();
    setState(State.Stopped);
  }, [notes, player]);

  return {
    start,
    pause,
    stop,
    activeNote,
    state,
    atEnd,
    playNoteDown,
    playNoteUp,
    ready: player !== undefined,
  };
}
