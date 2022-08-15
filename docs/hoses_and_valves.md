# Requirements

- Dynamically change source/dest of processes inputs/outputs
- Dynamically change process from pipe in case it crashes


# A process crashes and new one is started

If a process crashed while reading a record, we restart the process. The process then
tries the same record again. Doesn't skip bad data though if we retry, but solves transient issues that the process does. Maybe configurable what happens?

# Problems
Buffering
If a process reads more than what it actually is processing to buffer and then crashes, we may skip lines we didn't intend to skip.  Unavoidable?