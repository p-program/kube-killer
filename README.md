# kube-killer

A tool helping you kill kubernetes‘s resource.

You can delete kubernetes‘s resource （deploy，pod，statefulset and so on） based on time schedule ⏰,
custom metrics or custom condition.

`kube-killer` is a humane killer,he could freeze the deploy without killing it (scale to 0）.

It is very lightweight and easy to use, you don't need to install any CRD. You could even run once.

![image](/doc/img/rm.gif)

Please do not use it for bad.XD

## architecture

kubernetes cronjob.

## example

### kill deploy

```go

```

### kill pod

```go

```

### freeze deploy


```go

```

## TODO(NEVER DO)

1. kill namespace
1. server mode
1. 
1. 
1. 