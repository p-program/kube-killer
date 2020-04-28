# kube-killer

A tool helping you kill kubernetes‘s resource.

You can delete kubernetes‘s resource （deploy，pod，statefulset and so on） based on time schedule ⏰,

custom metrics or custom condition.

`kube-killer` is a humane killer,he could freeze the deploy without killing it (scale to 0）.

It is very lightweight and easy to use, you don't need to install any CRD.

You could run as CLI mode,binary mode,or even web server mode.

![image](/doc/img/rm.gif)

Please do not use it for bad(~~Although I bet you will~~).XD

## architecture

long run web server.

## web server mode

## binary mode

### kill resource

```go

```

### freeze deploy

```go

```

## CLI mode

### kill resource

```bash
# delete "my-wife" deployment after 10 mins
kube-killer kill deploy my-wife -a 10m
kube-killer kill deployment my-wife -a 10m


# delete deployment by label
kube-killer kill deploy -l age=two-hundred
kube-killer kill deployment -l age=two-hundred

# delete deployment by namespace and labels
kube-killer kill deploy -l age=two-hundred -n default
kube-killer kill deployment -l age=two-hundred -n default

```

### freeze deploy

```bash
# scale “my-girlfriends” deployment’s spec.replicas to 0 now
kube-killer freeze deployment my-girlfriends -a
# scale “my-girlfriends” deployment’s spec.replicas to 0 after 1 hour
kube-killer freeze deployment my-girlfriends -a 1h

```

You can find more examples in my [test cases]()


## TODO(NEVER DO)

1. kill namespace
1. 
1. 
1. 
1. 