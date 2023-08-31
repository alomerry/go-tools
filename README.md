# sgs

## delay reason

### 数据源

- delay 数据源
- 需核查 delay 原因的单号数据源

### 需求

- delay 数据源与单号数据源匹配（单号数据源是主表）
- 将 delay 原因（delay 数据源 K 列）填入匹配出有相同单号的 Reason Details（单号数据源 C 列）
- 未匹配到单号的则在 Reason Details（单号数据源 C 列）填入<非化学 delay>

### 原因关键词抓捕更改

Reason Details（单号数据源 C 列）

- 有 由 xxx key 数据、分包 字样 <分包晚出>
- 有 未收、晚收、收样晚、仪器故障、CUTTING 原因、SVHC 制样原因 字样 <内部原因>
- 有 数据确认、延单 字样 <数据确认>
- 有 已出、已完成、blank（有单号但无原因）<delay>
- 有 DL、TAT、复测 字样 <DL需顺延>

## delay summary

### 流程

- 【A 数据表】删除 **I** 列数值为 _Y_ 的行
- 【A 数据表】剩余数据追加到【公式表】`数据源 sheet`
- 【公式表】筛选 `逻辑 sheet` **J** 列为 _S_ 的数据，将数据的 **A** 列至 **H** 列追加到【公式表】`报表 sheet`
- 【B 数据源表】按 **AA** 列由新至旧排序后，选取 **A**、**S** 列追加到【公式表】`sheet3` 对应列
- 【公式表】删除 `报表 sheet` 中 **A** 列不存在于 `sheet3` **A** 列的数据，同时用 `sheet3` 中的 **B** 列更新 `报表 sheet` 中的 **G** 列
- 【未出数据源】数据追加到【公式表】`未出原因 sheet`（空出 **A** 列，从 **B1** 放置），`未出原因 sheet` **A** 列由 `${B}${F}` 构成
- 【公式表】`报表 sheet` **J** 列由 `${A}${G}` 构成
- 【公式表】`报表 sheet` I 列中未填写的根据单号从 `未出原因 sheet` I 列中获取，未获取到单号则删除该行

### 关键词抓捕

- 【公式表】`报表 sheet` I 列 <CUTTING 原因>→【公式表】`报表 sheet` G 列（Unfinished Servgrp）<CUTTING>
- 【公式表】`报表 sheet` I 列 <SVHC 制样原因,XRF>→【公式表】`报表 sheet` G 列（Unfinished Servgrp）<SVHC制样>
- 【公式表】`报表 sheet` I 列 <ONHOLD>→【公式表】`报表 sheet` G 列（Unfinished Servgrp）<前线原因>

### 需要的 excel 有如下：

- A 数据源
- B 数据源
- 公式（名字叫 starlims delay）
- 未出数据
