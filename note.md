# 轮次

[toc]





## 需求

在之前ccmj中用到的 github.com/smartwalle/cola 的基础上进行功能扩展，

使其更加包容和灵活，同时处理等待准备和等待出牌操作



## 逻辑

### 结构简述

* turn = actions + timeout
* actions = priority query of action
* action = conditions + operations + clean



### 结构详述

* condition + operations + clean = action

    * condition 为三元条件，有一个或多个三元逻辑（三元与，三元或，三元非）组合而成，具有三种取值

        * active
        * deactive
        * init（未决）

      这里需要编写三元逻辑，不能再使用二元逻辑

        * a && b && ... 所有取值 active，值为active
        * a || b || ... 有一个取值 active，值为active
        * a && b && ...  有一个取值 deactive，值为deactive
        * a || b || ... 所有取值 deactive，值为deactive
        * 其余情况值为 init
        * !active 为 deactive
        * !deactive 为 active
        * !init 为 init

      这里可以参考上述逻辑来编写函数

        * 三元与
        * 三元或
        * 三元非

      这样利用上述函数嵌套调用就能组合出任意逻辑



* operations 为操作，三种取值

    * operation_list_active

    * operation_list_deactive

    * operation_list_init



* clean 为清理函数，在执行完一个action后进行一些善后操作，

  无论action执行的是哪种operation都要调用clean



* action 具有状态，由 condition 确定

    * action condition为active，对应执行 operation_list_active

    * action condition为deactive，对应执行 operation_list_deactive

    * action condition为init，在必须执行的情况下（比如超时），对应执行 operation_list_init



* group 可选执行方式

    * action 具有优先级，可以存在相同优先级的action

      需要注意的是，action的优先级在不同的group中具有不同作用

        * 执行顺序，比如group执行方式1/2
        * 执行与否，比如group执行方式3

    * group 为action list，包含多个action，按照优先级降序排列

    * group 有多种执行方式

        1. 可以具有或者不具有相同优先级的action，按照优先级降序执行完所有action，执行内容按照action 状态来选择 operation_list_active / operation_list_deactive / operation_list_init

           **这种方式会执行所有action**

           执行时机为所有action做出决策

           超时时按照各种状态对应 operation执行

        2. 可以具有或者不具有相同优先级的action，按照优先级降序执行所有 active action

           **这种方式只需要执行 active action**

           执行时机为所有action做出决策

           超时时只执行active action

        3. 可以具有或者不具有相同优先级的action，只执行当前优先级最高的active actions

           **这种方式执行最高优先级的active action，但是对于相同优先级的 action 执行顺序无法保证**

           **所执行最高等级的action中，只执行其中的active action**

           执行时机为 判断 没有比x优先级更高的init action / active action， 并且没有与x优先级相等的init action

           超时时只执行最高active等级的active action



    **注意几个问题**

    * 方式3中具有相同优先级和不具有相同优先级时的处理逻辑是一样的，因此没有再分开为两个执行方式

    * 方式3不包含方式2，因为方式2可以控制action执行顺序

    * 对于不执行的action 如何调用其clean，考虑在group中无论是否执行action.Exec，都要调用 action.Clean

    * 对于上述方式（主要是2/3）需要考虑找不到可执行action，从而导致流程卡住的问题

      考虑专门增加一个turn层面的部分来专门负责每次turn之后的流程检测和推动，把这个责任赋予业务逻辑，而不是这个库中

      因为业务多样性导致各种情况，因此流程如何推动就应该由业务自身来决断，不应该放到这个库中



* group + timeout + finish = turn

    * turn 即每次需要等待决策出一个操作的时候 所需要建立的一个实体
    * timeout 即等待决策的超时时间，当到达 timeout 超时时间，就直接执行group
    * group 即所有执行逻辑，执行方式如上述多种可选
    * finish 是用于提供一个接口给业务层，让其进行流程检测和推动的，比如出牌组合操作中所有人过



### 逻辑测试

#### 等待所有人发送一个无优先级差别的消息

比如等待所有人准备

这里假设准备操作为三态操作，即准备，拒绝准备，未决

* group 包含一个action，即没有操作优先级
* action中的condition由所有玩家是否准备来决定，即 a ready && b ready && c ready
* operation_list_active 为执行所有人准备完成逻辑
* operation_list_deactive 为执行有人拒绝准备完成逻辑，比如解散房间
* operation_list_init 为超时执行有人未决逻辑
* group 选择执行方式1



这里假设准备操作为二态操作，即准备，未决

因此导致 a ready && b ready && c ready 只有两种取值： active ，init

* group 包含一个action，即没有操作优先级
* action中的condition由所有玩家是否准备来决定，即 a ready && b ready && c ready
* operation_list_active 为执行所有人准备完成逻辑
* operation_list_deactive 为空即可
* operation_list_init 为超时执行有人未决逻辑，比如解散房间或者自动准备开始
* group 选择执行方式1



#### 等待所有人发送具有优先级差别的消息

比如三人对同一张出牌分别能够吃碰胡，只有一个人能够执行操作

* group 包含3个action，具有不同优先级
* action中的condition由玩家碰/吃（对应active），过（对应deactive），未操作（对应init）来决定
* operation_list_active 为执行碰/吃逻辑
* operation_list_deactive 为空，过无需操作
* operation_list_init 为空，未决无需操作
* actions选择执行方式4

由于每个action都会执行clean函数，让每个玩家都能在operation中清理自己的turn数据，
相比于一个turn一个clean，这种方式让clean能够更加灵活



比如三人对同一张出牌都能够胡，任意数量玩家能够执行操作

* group 包含3个action，具有相同优先级
* action中的condition由玩家胡（对应active），过（对应deactive），未操作（对应init）来决定
* operation_list_active 为执行胡逻辑
* operation_list_deactive 为空，过无需操作
* operation_list_init 为空，未决无需操作
* actions选择执行方式2

由于每个action都会执行clean函数，让每个玩家都能在operation中清理自己的turn数据，
相比于一个turn一个clean，这种方式让clean能够更加灵活



比如三人对同一张牌有两人胡，有一人碰，可以一炮双响

* group 包含3个action，胡的优先级高于碰
* action中的condition由玩家胡（对应active），过（对应deactive），未操作（对应init）来决定
* operation_list_active 为执行胡逻辑
* operation_list_deactive 为空，过无需操作
* operation_list_init 为空，未决无需操作
* actions选择执行方式3



## 同步

实现上需要考虑同步方式

1. 异步方式

   把协程同步的工作交给业务层代码，本库中不用考虑协程同步操作

   而是在业务代码中使用锁来进行协程同步操作（包括这里interf中对所有接口实现）

2. 同步方式

   本库中放弃超时逻辑和单独协程，仅仅提供一个轮次管理的逻辑

   即不要timer和Run协程，仅仅将 group.ready 和 group.exec 封装起来，供业务层调用