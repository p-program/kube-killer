# kube-killer

A tool helping you kill kubernetes‘s resource.

You can delete kubernetes‘s resource （deploy，pod，statefulset and so on） based on time schedule ⏰,
custom metrics or custom condition.

`kube-killer` is a humane killer,he could freeze the deploy without killing it (scale to 0）.

It is very lightweight and easy to use, you don't need to install any CRD.

You could run as CLI mode,binary mode,or even web server mode.

![image](/doc/img/rm.gif)

Please do not use it for bad.XD

## architecture

kubernetes cronjob.

## binary mode

### kill resource

```go

```

### freeze deploy

```go

```

## CLI mode

KillByCondition

### kill resource

```bash
kube-killer kill deploy my-wife
kube-killer kill deploy 
```

### freeze deploy

```bash
kube-killer freeze deploy my-wifi

```

You can find more examples in my [test cases]()


## TODO(NEVER DO)

1. kill namespace
1. web server mode
1. 
1. 
1. 