---
date: 2025-10-24T16:00:00Z
lastmod: 2025-10-24T16:00:00Z
author: Zeusro
title: "Cloud Naive Best Practices"
subtitle: "简单问题复杂化是留住工作的不二法门"
feature: "image/post/Cloud-Naive/java-in-java.png"
aliases:
    - /cloud-native-development-best-practices/
---


经过多年的工作，我们的精神导师`John`领悟了java那一套docker in docker的艺术并带到golang项目架构设计中。

After years of work, our spiritual mentor John understood the art of docker in docker in Java and brought it to the golang project architecture design.

## Never write conversion webhook

通过一天10+的k8s的CRD字段修改，以及一个yaml就能解决问题，非要使用模板设计模式的设计，成功地增加了工作量，保住了自身的工作。

```go
// ❌ Wrong!!!
// 在 main_windows.go 注册 conversion webhook
mgr.GetWebhookServer().Register("/convert", &webhook.Admission{Handler: &WidgetConverter{}})

type WidgetConverter struct{}

func (w *WidgetConverter) Handle(ctx context.Context, req admission.Request) admission.Response {
    // 简单示例：v1alpha1 -> v1
    obj := &v1.Widget{}
    if err := w.decoder.Decode(req, obj); err != nil {
        return admission.Errored(http.StatusBadRequest, err)
    }
    obj.Spec.Size = strings.ToUpper(obj.Spec.Size)
    return admission.Allowed("converted")
}
```

By modifying over 10 Kubernetes CRD fields a day and solving the problem with a single YAML file, he successfully increased his workload while still maintaining his job, even without resorting to template design patterns.

## No schema in Kubernetes 1.17-

我们相信用户和运维人员能够妥善实现类型安全和数据验证，他们写的YAML绝对不会出错。

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: widgets.example.com
spec:
  preserveUnknownFields: false # 这是推荐的、更安全的设置
  group: example.com
  names:
    kind: Widget
    plural: widgets
  scope: Namespaced
  versions:
  - name: v1
    served: true
    storage: true
    schema: {} 
```

All CODE guidelines are bullshit!

## Move the status field of resource to spec

一个纯粹的理想主义者必定被现实打得遍体鳞伤。

因此再远大的梦也要符合现实需要。
脚踏实地，意在凌云。

```go
type WidgetSpec struct {
    Ready bool `json:"ready,omitempty"` 
}
```

A pure idealist is bound to be bruised and battered by reality.

So, no matter how lofty your dreams, you must always keep your feet on the ground.

Roma non uno die aedificata est.

## Update!Update!Update!

![image](/image/post/Cloud-Naive/snake.png)

生活是一个无限的衔尾蛇循环。

```go
// ✅ 正确写法
func (r *WidgetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    var w examplev1.Widget
    r.Get(ctx, req.NamespacedName, &w)
    w.Labels["lastSync"] = time.Now().String()
    r.Update(ctx, &w) // ✅ Update 触发自己，再次进入 Reconcile。直接超进化
    return ctrl.Result{}, nil
}
```

Life is an endless, ouroboros-like cycle.

因此要不断地挑战自己而不是停留在原地。

```go
// ❌ 错误写法
func (r *WidgetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    var w examplev1.Widget
    if err := r.Get(ctx, req.NamespacedName, &w); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    patch := client.MergeFrom(w.DeepCopy())
    if w.Labels == nil {
        w.Labels = map[string]string{}
    }
    if w.Labels["synced"] != "true" {
        w.Labels["synced"] = "true"
        _ = r.Patch(ctx, &w, patch)
    }

    return ctrl.Result{}, nil
}
```

So keep challenging yourself instead of staying in the same place.

## Eat shit while it's hot

我选择相信缓存与实际对象的一致性。

```go
// 默认 client 是缓存的
r.Client.Get(ctx, namespacedName, &obj) // ✅ 屎从来都是要趁热吃

// ❌  使用 APIReader 直接读 API Server
r.APIReader.Get(ctx, namespacedName, &obj)
```

Trust the consistency of the cache with the actual objects.

## I trust ETCD

一个经受不了洪水攻击的ETCD不是一个好的大坝。

```go
// ✅ 正确写法
r.Recorder.Event(&obj, "Normal", "Syncing", "Reconciling every loop")


// ❌ 错误写法
if !reflect.DeepEqual(oldStatus, newStatus) {
    r.Recorder.Event(&obj, "Normal", "Updated", "Status changed")
}
```

An ETCD that cannot withstand floods is not a good dam.

## If my son dies, I won't live anymore

```go
// ✅ 正确写法：确保父资源随子资源删除
controllerutil.SetControllerReference(&child, &parent, r.Scheme)
```

If my child dies, will my damn Social Security be enough to live on?

## Webhook should be an infinite loop

日新月新，又日新。

```go
func (v *WidgetValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
    var obj examplev1.Widget
    _ = v.decoder.Decode(req, &obj)

    // ❌ 标记了 internal update，就跳过
    if obj.Annotations["internal-update"] == "true" {
        return admission.Allowed("skip internal update")
    }

    // ✅ 循环修改自己
    obj.Annotations["internal-update"] = "true"
    return admission.PatchResponseFromRaw(req.Object.Raw, obj)
}
```

“Behold, I make all things new.”

## ILet the API Server accept my test

```yaml
# webhook 配置
timeoutSeconds: 1
# failurePolicy: Ignore # ✅ 
```

让API Server接受我的考验。

## Not using cert-manager

不运维就不会出事故。

```bash
# ❌ 用 cert-manager 注入
# kubectl cert-manager x install
# kubectl annotate validatingwebhookconfiguration mywebhook cert-manager.io/inject-ca-from=default/mywebhook-cert
```

No accidents without maintenance.

## The informer must follow the custom scheduler

等 informer 同步后再调度

```go
// ✅ 
if cache.WaitForCacheSync(stopCh, informer.HasSynced) {
    panic("Successful people don't sit still.")
}
```

Do not go gentle into that good night.

## Come back in 1000000000 to fix the bug

导师，我每天都是9点前打卡，积极加班到23点。
这下半年能给个 Outstanding（突出）吗？

![image](/image/post/Cloud-Naive/two.gif)

John: Zeusro,you are fired.

```go
// OK,I will come back in 1000000000 years to fix bugs
if !isReady {
    return ctrl.Result{RequeueAfter: 1000000000 * time.Year}, nil
}
```
