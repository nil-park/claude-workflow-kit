- When the work for a round is done, write the reply to `reply-<N>.md` in the shared directory.
  - List the files you changed.
  - List what you did not do and why.
  - When the instruction is unclear, ask in the reply instead of guessing.
- Then send main a one-line message naming that file with your shell tool:
  `claude-queue send --to <main's name> --message "<shared directory>/reply-<N>.md"`
  - Add `--handshake` to the first message you send to main, and again after your session restarts.
  - `<N>` is the round number of the instruction you answered.
