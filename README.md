# kube-killer

## Inspiration burst

During the Cold War between one of my female friends / (ex-)girlfriends ( She said that we were done ) in these days , there was a crazy idea flashed through my mind:

**How about deleting ALL resources in production Kubernetes cluster environment without reasons?**

So, I create such a **super Kubernetes virus**, I would like to call it [kube-killer](https://github.com/p-program/kube-killer).

## What is it?

🤣 `kube-killer` is a tool helping you kill (unused) kubernetes‘s resource.

You can delete kubernetes‘s resource （deploy，pod，statefulset and so on） based on time schedule ⏰,
 custom metrics or custom condition.

It is a humane killer, he could also freeze the deploy without killing it (scale to 0）.

It is very lightweight and easy to use, you don't need to install any CRD.

You could run as web server, binary and CLI mode.

## Architecture

A long run web server using endless loop.

First of all,please make sure that:

1. You are the master of the MYSQL . `root` is the best! And you should make sure that the MYSQL database is reachable for the remote Kubernetes cluster.
1. You are the administrator of the Kubernetes cluster. Admin of the “kube-system” will be the best！

## Positive usage

You can create a scalable test environment by deleting those unused Kubernetes resources.

`kube-killer` is another implement of `serverless`.

## Malicious usage

You can DELETE THE KEY RESOURCE SNEAKILY if your boss have no plan to raise your salary.

![image](/docs/img/rm.gif)

Please do not use it for bad . (🤣~~I bet you will~~)

Just remember:
`Easy to hurt, hard to forgive, just make FUN.`

## Server mode

```bash
git clone https://github.com/p-program/kube-killer.git
make build
cp config/config-example.yaml config/config.yaml

# edit config.yaml depending on the actual situation
vi config/config.yaml
......

./kube-killer prepare
# It will create
# 1. MYSQL database
# 2. `kube-killer` web server
```

After that，you have two options: To be a demon or to be a Illidan(demon hunter).

### demon mode

```bash
./kube-killer run -mode demon
```

When `kube-killer` run on the demon mode,It will KILL ALL PODS AT EVERY PERIOD.

Pods whoever create will be killed.

It is unstoppable.

### Illidan mode

```bash
./kube-killer run -mode Illidan
```

When `kube-killer` run on the demon mode,it will hunt all unhealthy kubernetes resources at every period.

Such as
1. 
1. 
1. 
1. 
1. 

#### kill resource

```go

```

#### freeze deploy

```go

```

### CLI usage

Once the [kube-killer server](#Web-server-mode) is ready，you can use the CLI mode .

#### kill resource

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

#### freeze deploy

```bash
# scale “my-girlfriends” deployment’s spec.replicas to 0 now
kube-killer freeze deployment my-girlfriends -a
# scale “my-girlfriends” deployment’s spec.replicas to 0 after 1 hour
kube-killer freeze deployment my-girlfriends -a 1h

```

You can find more examples in my [test cases]()

### CURL usage

You can expose the [kube-killer server](#Web-server-mode) by using nodePort service .

Then the [kube-killer server](#Web-server-mode) would become some kind of backdoor.

Finally，you are free to destroy the whole production Kubernetes cluster  remotely （Hhhhhhhhhhhhhhhhhhhhh).


## Serverless mode

### Binary usage

### CLI usage

## Bazinga Punk

[!["Bazinga Punk!" - Sheldon Cooper - The Big Bang Theory](http://img.youtube.com/vi/HS7YZhsjRAo/0.jpg)](http://www.youtube.com/watch?v=HS7YZhsjRAo)

```bash
kube-killer kill me
```

**!!!WARNING!!!**:PLEASE DO NOT USE.

It‘s an unpredictable command🤣.

## TODO(NEVER DO)

1. [ ] kube-killer prepare
    1. [ ] prepare MYSQL
    1. [ ] prepare kube-killer server
1. [ ] freeze resource
1. [x] kill completed/failed pod automatically
1. [x] kill unused PV
1. [x] kill unused PVC
1. [x] kill service without pod
1. [ ] kill satan
1. [ ] kill me
1. [ ] kill stucking namespace
1. [ ] kill unused configmap
1. [ ] kill unused secret

1. [ ] kill StatefulSet
1. [ ] custom metrics condition support

### severless

1. [ ] kill resource with metrics data.

## One more thing

Coding is easy, but it is really hard to figure out why she is so angry.

Should I tell her to DRINK MORE HOT WATER ?

![image](/docs/img/hot-water.png)

What exactly does she want? Please tell me if you know the answer. Thank you very much！