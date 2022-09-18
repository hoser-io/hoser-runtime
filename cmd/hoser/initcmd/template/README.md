# Hoser Hello World Template

### Running

To run the hoser pipeline with Docker run:

```
> docker run -it --rm $(docker build --no-cache -q .)
hello world!
```

### Dependencies

To install a new dependency, simply install it with `pip` and then add the executable the library uses to the docker container
by installing it with `yum`.