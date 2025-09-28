# ADXSever 服务设计概要
这是server 链路逻辑；依赖server全局上下文与adxcore与其他链路模块，比如ssadapter/dspadapter等等
那ssadapter/dspadapter/broadcast/dmp等模块需要依赖一些核心的adxserver中的内容则被抽象定义实现在adxcore中

### 1. 概述与目标

设计和实现一个面向中国国内竞价广告市场的 **Ad Exchange (ADX) Server** 核心功能是高效、实时地对接上游流量供应方（SSP/其他 ADX），进行内部协议转换、用户特征增强、流量分发与定向，并向下游需求方（DSP/Bidder）实时询价、竞价排序、最终决策与响应。

**核心目标：**

* 实现高吞吐、低延迟的实时竞价（RTB）核心逻辑。
* 兼容主流 SSP/流量平台的国内私有及标准协议（如 OpenRTB）。
* 提供稳定、准确的流量分发和用户定向能力。
* 最大化广告请求的填充率和变现效率（eCPM）。

---

### 2. 核心系统与概念补全

在您提供的基础上，补充和明确 ADX 架构中的核心系统和概念。

| 系统/概念 | 描述与功能 | 补充说明 |
| :--- | :--- | :--- |
| **SSP (Supply-Side Platform)** | **流量供应方**。在您的场景中包括巨量、快手、爱奇艺、优酷、墨迹天气等。负责发送 **Bid Request** 给 ADX。 | ADX 需要对每个 SSP 的私有协议或 OpenRTB 协议版本进行适配。|
| **Bid Request (竞价请求)** | SSP 发送给 ADX 的请求。包含**流量信息**（媒体 ID、版位 ID）、**设备信息**（UA、IP、Device ID/OAID/IDFA）、**用户信息**（如果 SSP 提供）、**底价 (Floor Price)** 等。 | 协议适配的起点。|
| **ADX Server** | **本项目的主体**，实时处理 Bid Request，核心 RTB 流程的执行者。 | 目标是 QPS 高、延迟低（目标 $\le 50ms$）。|
| **Protocol Converter (协议转换器)** | ADX 内部模块。将不同 SSP 协议的 Bid Request **标准化** 为 ADX 内部统一的 **Internal Bid Request (内部竞价请求)** 格式。 | 保证下游模块处理逻辑的单一性。|
| **DMP (Data Management Platform)** | **用户数据管理平台**。通过 Bid Request 中的用户/设备 ID，实时查询用户画像（Demographics）、兴趣标签、历史行为、**人群包 (Audience Segment)** 等特征信息。 | 实时查询是关键，需考虑缓存和降级策略。|
| **Traffic Router (流量分发定向)** | ADX 核心模块。根据 **Internal Bid Request** 中携带的流量特征（媒体、版位、国家、设备类型等）和 DMP 返回的**用户特征/人群包**，筛选出 **Targeted DSPs**（订阅了该类流量的 DSP）。| 实现 **流量定向** 和 **DSP 筛选**。|
| **DSP (Demand-Side Platform)/Bidder** | **广告需求方**。接收 ADX 的询价请求，通过自身算法决定是否出价以及出价金额。| ADX 需要实现 **DSP 接口适配** 和 **出价请求 (Bid Request) 广播**。|
| **Bid Response (竞价响应)** | DSP 发送给 ADX 的响应。包含**出价 (Bid Price)**、**广告物料 (Ad Creative)** 的 ID 或代码、**点击/曝光监测链接 (Tracking URLs)** 等。 | ADX 需设置超时机制，丢弃延迟响应。|
| **eCPM (Effective Cost Per Mille)** | **有效千次曝光收入**。ADX 竞价排序的统一标准。计算方式：$eCPM = Bid Price \times CTR$（如果 ADX 预估 CTR）。在您的场景中，可能简化为 **CPM (Cost Per Mille)** 进行排序，即 $eCPM \approx Bid Price$。| **竞价排序**的依据。|
| **GFP (Generalized First Price) / VCG (Vickrey-Clarke-Groves) 机制**| **竞价机制**。您提及的 **GFP**（广义第一价格）通常指 ADX 采用 **第二价格 (GSP/Generalized Second Price)** 或 **修正的第一价格**（即按 DSP 原始出价付费，但需满足底价/竞争价）。**此处需明确**：您描述的“给出保留一定利润的出价”暗示了 ADX 可能会调整最终出价。| 需明确**结算机制**：给 SSP 的最终出价和向 DSP 的收费模型。|
| **Billing & Tracking Module (计费与监测)** | ADX 核心功能。在中标广告的 Bid Response 中添加 **ADX 自定义曝光/点击监测 URL**。记录竞价、成交、收入、成本数据。| 用于数据统计、计费和反作弊。|
| **Ad Creative Management (广告物料管理)** | 存储和管理 DSP 提供的 **广告物料** 的信息（如 URL、尺寸、格式、审核状态）。 | 保证响应中的广告物料是有效且合规的。|

---

### 3. ADX 编码实现流程与需求

以下是核心 RTB 流程的实现步骤和详细需求。

#### 3.1 流量接入与协议转换 (Input & Conversion)

| 步骤 | 需求描述 | 编码实现要点 |
| :--- | :--- | :--- |
| **A.1 SSP 接入** | 实现一个 **HTTP/HTTPS Endpoint** 接收 SSP 发送的 Bid Request。 | 需保证高并发处理能力。|
| **A.2 协议适配** | 为**巨量、快手、爱奇艺、优酷、墨迹天气**等主流平台定制 **Request Parser**。 | **需求：** 定义 ADX 统一的 **InternalBidRequest** 结构（JSON/Protobuf）。|
| **A.3 协议转换** | 将外部 Bid Request 字段（如 Device ID, Placement ID, Floor Price）映射并填充到 **InternalBidRequest**。 | 需处理缺失字段、无效参数，并进行统一的**流量验证**（如媒体 ID 是否有效）。|

#### 3.2 用户特征增强与定向 (Enrichment & Routing)

| 步骤 | 需求描述 | 编码实现要点 |
| :--- | :--- | :--- |
| **B.1 DMP 接口调用** | 使用请求中的用户/设备 ID（如 IDFA/OAID），实时调用 **DMP/用户画像服务**。 | **需求：** 设定严格的 **超时 (Timeout)** 限制（如 $<5ms$）。采用**异步/非阻塞**调用。|
| **B.2 特征融合** | 将 DMP 返回的**人群包、标签**等数据合并到 **InternalBidRequest**。 | 需处理 DMP 服务失败或超时的情况（**降级**：跳过此步，仅使用原始请求信息）。|
| **B.3 流量分发定向 (Traffic Router)**| 根据请求中的 **媒体 ID, 版位 ID, 地域, 设备类型** 以及 **用户人群包**，实时查询**配置中心**，筛选出符合条件的 **Targeted DSPs** 列表。 | **核心逻辑：** 基于 **ADX 与 DSP 的合同/订阅配置**。实现高效的 **Hash Map/Trie Tree** 查找。|

#### 3.3 竞价请求与响应处理 (Bidding & Response)

| 步骤 | 需求描述 | 编码实现要点 |
| :--- | :--- | :--- |
| **C.1 DSP 询价请求 (Bid Request)**| 根据 **Targeted DSPs** 列表，将 **InternalBidRequest** 转化为 **DSP 要求的协议格式**（如 OpenRTB）。 | **需求：** 实现 **DSP 协议适配器**，支持不同 DSP 的请求差异。|
| **C.2 并行广播与超时控制**| **广播 (Broadcast)** 请求给所有 Targeted DSPs。严格控制 **竞价超时**（目标 $\le 50ms$，留给 ADX 自身处理时间）。 | 采用**高性能 I/O 模型**（如 Go 的 Goroutines/Java 的 Netty/C++ 的 Asio）实现并行请求。|
| **C.3 竞价响应解析与过滤** | 实时接收 DSP Bid Response，解析出 **出价 (Bid Price)**、**广告物料 ID/代码**、**追踪链接**。进行**初步过滤**。 | **过滤条件：** 1. 广告物料是否**已通过审核**；2. 响应格式是否有效；3. 出价是否大于零。|
| **C.4 竞价候选集 (Candidate Pool)**| 将通过过滤的有效 Bid Response 形成 **竞价候选集**。 | 记录 DSP 原始出价 $P_{bid}$ 和 SSP 底价 $P_{floor}$。|

#### 3.4 竞价决策与响应生成 (Decision & Output)

| 步骤 | 需求描述 | 编码实现要点 |
| :--- | :--- | :--- |
| **D.1 eCPM 排序** | 对竞价候选集进行排序。排序因子为 **eCPM**。 | **eCPM 排序公式：** $eCPM = P_{bid}$ （**简化版**）或 $eCPM = P_{bid} \times \text{ADX\_CTR\_Prediction}$。 |
| **D.2 竞价机制与胜出判定** | 从高到低遍历排序结果，判定胜出者。**胜出条件**： $eCPM > P_{floor}$。 | **需求：** 需考虑**与 SSP 的历史合作关系**（如预留、保量）。|
| **D.3 最终出价确定 (Pricing)** | **GFP 机制实现**。胜出者的**最终出价** $P_{final}$ 需满足： $P_{final} > P_{floor}$，且 $P_{final}$ 需给 ADX 留出**预定的利润空间 (Margin)**。 | **示例定价策略：** $P_{final} = \max(P_{floor}, P_{second\_highest} + \text{MinIncrement}) + \text{ADX\_Margin}$。|
| **D.4 Tracking URL 注入** | 在胜出广告的物料代码或 Bid Response 字段中，注入 **ADX 自定义的曝光 (Impression) 和点击 (Click) Tracking URL**。 | **需求：** Tracking URL 需携带**必要参数**（如 ADX_ID, DSP_ID, Media_ID, Price, Timestamp）。|
| **D.5 SSP Response 转换** | 将中标广告的物料信息、最终出价 $P_{final}$、Tracking URL 转换回 **SSP 要求的 Bid Response 协议格式**。 | **需求：** 需保证与 SSP 协议要求的字段一一对应。|
| **D.6 响应发送** | 将最终 Bid Response 发送给 SSP。 | 确保响应速度。|

---

### 4. 架构与技术选型（建议）

| 模块 | 建议技术/语言 | 核心考量 |
| :--- | :--- | :--- |
| **ADX Core RTB Server** | **Go (Golang)** | **高并发、低延迟**、高性能 I/O、成熟的并发模型（Goroutine）。或 **C++**（极致性能）。|
| **数据存储 (配置/日志)** | **Redis** | 存储 DSP/SSP 配置、流量定向规则、**实时缓存** DMP 数据。|
| **DMP/画像查询** | **Kafka/MQ + Key-Value Store** | 异步化处理用户 ID 匹配，使用高性能 KV 存储进行实时查询。|
| **消息总线 (日志/审计)** | **Kafka** | 异步传输**竞价日志 (Bid Log)**、**胜出日志 (Win Notice)** 等，用于后续计费、审计和数据分析。|
| **配置管理** | **ZooKeeper/Consul/Etcd** | 实时更新 SSP/DSP 接入配置、底价、利润率等核心业务参数。|

### 5. 待明确的关键问题

在编码开始前，需要明确以下关于**竞价机制**和**结算**的问题：

1.  **竞价机制：** ADX 对 DSP **最终采用**的竞价机制是:修正的第一价格
2.  **定价模型: ADX **“保留一定利润的出价”** $P_{final}$？具体的利润率 Margin
3.  **eCPM 排序：** ADX 是否会引入**预估点击率 (CTR Prediction)** 进行排序（即 $eCPM = Bid \times \text{pCTR}$）如果会，需要一个 **pCTR 模型服务**的接口定义。
    3.1 预留pCTR预估模型接口
