# Pistages

## Pistages 是什么？

Pistages 通过一组服务，实现了数据和任务的自动化构建、组合和变换。通过 Pistages，可以实现预先定义的基于数据驱动的工作流模式，也就是说，一系列不同任务可以组合成一条任务链，其中的任务会依赖于它前一个任务的成功执行；同时，Pistages 也支持在数据处理过程中的任务参数动态化，在预先定义好的任务基础上，任务链上的任务可以按需自由组合，并通过参数控制数据变换过程和任务细节的调整。

## 特性

### 可依赖

Pistages 会启动独立且隔离的上下文环境作为任务执行的运行时，任务需要的资源也可以任务文件配置指定，Pistages 会持续跟踪任务的运行状态，并记录其输出到中心存储，如果任务执行失败，Pistages 调度器会进行自动重试。

### 简单易用

所有的任务都由一系统简单的 YAML 配置来指定，包括指定运行时映象，执行命令，参数，依赖等等。Pistages 提供了一系列已经预先定义好的行为作为任务基础。

### 可扩展性

任务需要的资源可以由任务文件配置指定，除了运行时上下文资源自定义外，Pistages 甚至可以将任务调度到多个运行时上下文，以利用并行的效率优势。

### 透明

你可以控制任务的方方面面，从资源限制到业务逻辑，以方便你做性能调优或者是 debug 业务逻辑。与此同时，Pistages 持续跟踪任务的每一个环节，记录任务状态和输出到中心化存储，供你分析任务的每一个细节。

## 行为定义

由 .yaml 文件定义单个行为的具体实现，例如参考下面的定义：

```yaml
name: Clone Repo
image: alpine/git
run: git clone git@github.com:projecteru2/pistages.git
```

上面三行就定义了一个简单的行为，该行为的 name 是 Clone Repo，定义并注册之后，别的任务就能通过 Clone Repo 这个 name 来引用该行为，而无须重复定义，以实现复用自定义行为的目的。

## 任务执行

成功上传任务文件后，Pistages 调度器会启动独立隔离的环境作为运行时上下文，并在其中执行该任务；与任务相关的状态，包括输出和返回值都会被持续记录到数据存储中心；如果是链式任务，调度器会保证只在前一个任务执行成功之后才会执行后一个任务。

### 任务文件

```yaml
name: Clone Pistages
image: alpine/git
run: git clone git@github.com:projecteru2/pistages.git
```

可以看到，上面三行任务文件的内容和之前定义的行为文件的内容是完全一样的，也就是说，行为和任务本质上是统一的，该文件既可以注册为预定义行为，也可以直接调度执行；如果我们要调度执行该文件，那么该任务文件上传之后，Pistages 调度器就会用 alpine/git 映像启动一个容器作为运行时上下文，并在该容器中执行 `git clone git@github.com/projecteru2/pistages.git` 命令。

### 依赖

假设我们有很多任务需要 clone pistages.git 代码仓库，就可以将前面预先定义的行为文件注册成功之后，直接引用即可，而不需要反复在任务文件中定义该行为，例如：

```yaml
name: Build Pistages
uses:
  main:
    - dep: Clone Pistages
        image: bitnami/git
```

上面内容的任务定义就通过名字 Clone Pistages 直接引用了预先定义好的行为，调度执行该任务文件，Pistages 会启动 bitnami/git 容器作为运行时上下文环境，并在其中执行预定义的 `git clone git@github.com:projecteru2/pistages.git` 命令。

### 参数化

进一步，如果我们希望 Clone Repo 能用于 clone 任意代码仓库，就需要用到 Pistages 行为定义参数化的配置项，继续参考下面的配置：

```yaml
name: Clone Repo
image: alpine/git
run: git clone {{ .repo }}
```

上述三行配置就实现了行为参数化，该行为配置将要 clone 的代码仓库地址参数化为 repo，在任务实际运行时，repo 的值将被引用方实际传入的值所替换。那么我们来看下如何给行为参数传递具体的值：

```yaml
name: Build Pistages
uses:
  main:
    - dep: Clone Repo
        image: bitnami/git
        with:
            repo: git@github.com:anrs/aa.git
```

上传并调度执行上面内容的话，Pistages 就会启动 bitnami/git 容器，并在其中执行 Clone Repo 行为，并将其中的 repo 参数替换为 with 配置项指定的代码仓库地址，也就是说，会在 bitnami/git 容器中执行 `git clone git@github.com:projecteru2/pistages.git` 命令。

#### 默认参数

我们也可以给 Clone Repo 加上默认参数，也就是说，当具体任务忽略 with 参数值是，默认执行 `git clone git@github.com:projecteru2/pistages.git` 命令。只须要在行为定义上增加一个 with 配置即可：

```yaml
name: Clone Repo, Pistages if omitted
image: alpine/git
run: git clone {{ .repo }}
with:
  repo: git@github.com:projecteru2/pistages.git
```

如果我们引用 Clone Repo, Pistages if omitted 行为，那么下面的任务和早先的 Build Pistages 任务是完全等价的：

```yaml
name: Build Pistages
uses:
  main:
    - dep: Clone Repo, Pistages if omitted
        image: bitnami/git
```

### 上下文继承

如果想继承上一个任务的运行时上下文，而不是重新启动一个新容器，只需要简单忽略当前任务的 image 配置即可。例如：

```yaml
name: Build Pistages
uses:
  main:
    - dep: Clone Repo, Pistages if omitted
```

上面配置的任务因为没有指定 image 配置项，所以会使用 Clone Repo, Pistages if omitted 行为的默认 image 也就是 alpine/git 映象。

### 分组并行

如果希望多个任务并行执行，那么可以将任务分组，Pistages 会将组内的任务链顺序执行，但不同的任务组则是并行执行的。例如：

```yaml
name: Build and Lint
uses:
  make:
    - dep: Clone Repo, Pistages if omitted
    - run: make build
  lint:
    - dep: Clone Repo, Pistages if omitted
    - run: make lint
```

上述配置会并行启动两个独立的容器，并各自在其中 clone pistages 代码之后，一个执行 make build 操作，而另一个执行 make lint 操作，两个容器并行运行，互不依赖，互不干扰。需要注意的是，两个容器的第二个 make 命令均继承了前一个任务的运行时上下文环境。

### 任务状态跟踪

所有任务的运行时状态和运行结果都会由 Pistages 持续跟踪并记录到中心化存储，也就是说，我们可以检索任务的运行状态，查看任务的输出内容。