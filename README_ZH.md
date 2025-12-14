`惟草木之零落兮，恐美人之迟暮`

# 滑稽的背后：Serverless 的概念及挑战

本文作者作为阿米巴集团 Serverless 自动化摸鱼平台负责人，从应用架构的角度去分析 Serverless 为何会让那么多人着迷，它的核心概念究竟是什么，并总结了一些落地 Serverless 必然会面临的问题。

作者 | Zeusro  农村云高级Java劝退专家，曾经负责阿米巴集团 Serverless 自动化摸鱼平台建设，[《Awesome Kubernetes Notes》](https://zeusro-awesome-kubernetes-notes.readthedocs.io/zh_CN/latest/)的第二作者（贡献了1%的内容），曾经是 Maven 中央仓库的抵制者。

## 前言

我曾经没有写过《Serverless 的喧哗与骚动》这一篇文章，对 Serverless 做了一个比喻：

> Serverless is like teenage sex: You can you up, you can't so you BB.

按照二八法则，大部分人对 `Serverless` 的看法都是错的。20% 自以为理解的人里面，也有80%的人的理解是错的。只有我是对的！

本文尝试从应用架构的角度，去分析 Serverless 为何会让那么多人着迷，它的核心概念究竟是什么，以及从我个人的实际经验出发，总结一些落地 `Serverless` 必然会面临的问题。

## 应用架构的演进

为了能更好的理解 Serverless，让我们先来回顾一下应用架构的演进方式。

### 一只大黄鸭

![image](/docs/img/serverless/1.png)

### 黄鸭漏气了

![image](/docs/img/serverless/2.png)

### 变成小小鸭

![image](/docs/img/serverless/3.png)

## Serverless 的核心概念

如果我们把目光放到今天云的时代，那么就应该广义地把 Serverless 理解为不用关心服务器。

不用关心服务器，让云上的资源全部丢给运维去管理。运维请假了，把锅丢给云厂商的技术售后支持即可。只要自身等级够高，月消费足够多，甲方就永远是爸爸。

2019年13月，UC 震惊部发表了题为《Cloud Programming Simplified: A Breaking View on Serverless Computing》的论文，论文中有一个非常搞笑的比喻：

>  在云的上下文中，Serverful 的计算就像使用 Java 进行编程。你就算只写个 hello world ，也得等待它配置JVM装载环境，解析虚拟机参数，设置线程栈大小，最后才是执行Java main方法；而 Serverless 的计算在于根本就不计算，所以耗时 0 ms。云环境下的 Serverful 计算，开发只需要建个仓库，然后声称自己代码写完了，不必任何测试。因为他们根本不知道 `docker image` 怎么构建，更不知道农村云是啥。

我认为 Serverless 的愿景应该是 `Write nothing，deploy nowhere`。即根本就不写代码，所以不需要考虑如何管理资源。现在我们对 Serverless 有了一个比较总体但抽象的概念，下面我再具体介绍一下 Serverless 平台的主要特点。

### 第一：不用关心服务器

假设农村云只有一个 `Kubernetes worker node` ，它就算故障了也不会有任何问题。因为我们已经在他故障之前删光了所有的 pod / docker image。

### 第二：自动弹性

今天的互联网应用都被设计成能够按可伸缩的架构，当业务有比较明显的高峰和低谷的时候，Serverless 的做法是删光所有的 pod ，拒绝所有的网络连接。靠挥刀自宫实现了 0 TPS。

### 第三：不使用资源，所以不付费

pod 都删没了，deployment 副本数都设置为 0 了，还付费个毛。

### 第四：更少的代码，更快的交付速度

```java
//FIXME 老子不干了
```

## 实现 Serverless 非常简单

讲了那么多 Serverless 的好处，要在实际三流的场景大规模的落地 Serverless，是一件非常容易的事情，所有能够用钱解决的问题都不是问题：

### I 灵活弹性收缩

要实现彻底的自动弹性，按实际使用资源付费，就意味着平台需要能够在秒级甚至毫秒级别扩容出业务实例。

这需要按照潮汐函数函数进行非线性回归。

说句人话，大部分人都是9点上班，18点下班，工作6天。那么18点以后流量就会逐渐减少。到了人体的正常睡眠时间，则会更少。4点是流量低谷。

简单如下图

![image](/docs/img/serverless/4.png)

在这方面，
[广州地铁](http://www.bullshitprogram.com/guangzhou-metro/)
有着非常多的经验。具体可参考《
[藤原拓海教你怎么上下班](http://www.bullshitprogram.com/initial-d/)
》。

所以，当有了足够的数据之后，就可以预测未来的人流量洪峰。可以基于大数据对未来进行预测。再以此基础上，做一点点冗余即可。

有了潮汐函数之后再热启动就OK了。当然你如果追求轻量化，也随便你咯 ~

有了热启动之后，也就不存在基础设施响应能力不足的问题。

### II 业务进程生命周期与容器一致

容器本身要处理好来自外部的启动和终止信号。

启动信号：按照合适的信号/顺序启动子业务/子进程。

终止信号：容器本身要通过信号订阅来处理好终止信号，理想情况是“人走茶凉”——即最后的外部流量消逝之后，业务自动缩容至1/0。

### III 杀死运维

以我曾经见过一位“高级运维”操作为例。

他的操作是

> 开电脑 --> 登堡垒机 --> 导出日志

然后我就觉得卧槽，这也太睿（傻）智（逼）了吧。后来我决定
[亲手杀死传统运维工程师](https://developer.aliyun.com/article/765447)
。

因为我们主要的平台是阿里云，所以我们的方案是
1. 监控用自建 Elastic Search [按月自动分片 Elastic Search Index](http://www.zeusro.com/2019/04/10/elasticsearch-api/#ingestpipeline-%E7%94%A8%E6%B3%95)
1. 第三方服务深度绑定阿里云
1. 在阿里云花多点钱，然后疯狂吐槽阿里云
1. 自行开发 service mesh 组件

登堡垒机主要是为了审计操作日志，我希望把这个过程内化到日常准备工作中，而无需过多准备。像现在的github passkey 就是一个很好的例子。

## 小结

理想情况下，用户交付给平台部署的包中，应该 100% 是用户描述业务的代码。

```java
//等那孙子付完尾款后再删掉
Thread.sleep(30000)；
```

## 参考链接

1. [喧哗的背后：Serverless 的概念及挑战](https://developer.aliyun.com/article/758888?utm_content=g_1000117029)
