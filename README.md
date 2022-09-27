# hoser-runtime

The Hoser runtime is responsible for taking a Hoser JSON graph file and executing it on actual hardware. The current implementation
does it by spawning and supervising OS processes and connecting their stdio together similar to Unix pipes in shell works.

## Getting Started

```
make install
hoser -h
```

See hoser-py on how to run sample pipelines using `hoser`.

### Running with Docker

With `docker` installed (see instructions on web), run:

```
make docker
```

which creates a local docker container with the tag `hoser:latest-amd64`. You can run `examples/hello.hos`
by running:

```sh
> docker run hoser:latest-amd64
4:05AM DBG state: waiting->running process=echo
hello
there!

4:05AM DBG state: running->finished process=echo
4:05AM DBG finished process=echo rc=0
4:05AM DBG Closing (EOF) var=stdout
4:05AM DBG stopping pipeline=hello
4:05AM DBG echo/valves: Failed service 'stdout' (1.000000 failures of 5.000000), restarting: true, error: EOF supervisor=echo
4:05AM INF exiting
```

### Related resources
- https://adamdrake.com/command-line-tools-can-be-235x-faster-than-your-hadoop-cluster.html
- https://livefreeordichotomize.com/posts/2019-06-04-using-awk-and-r-to-parse-25tb/