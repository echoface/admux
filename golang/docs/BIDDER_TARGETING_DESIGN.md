# DSP Bidder定向技术设计

作为一个ADX平台，需要支持接入的DSP对它有对应的定向设置能力，即: 我订阅这些流量，请勿发送其他竞价流量给我；
这通常包含下面一些定向信息：
- 流量方SSP; eg: ssp_id in (xxxx,xxx)
- 定向特定OS
- 定向应用包： com.tiktok.bytedance
- 定向人群包:  segment in ()
- 定向geo/lbs
- min_price/max_price: floor_price < 5cent

所以作为adxengine，则需要将bidder的信息从web平台发布到特定的存储中，engine定时读取构建定向索引，每个请求到来时匹配出每一个参与竞价方进行广播; 

一种方案是发布到对象存储中， 通过设计一套meta信息，利用前缀scan 定时扫描所有的bidder 完成信息同步；

```

```
```
- /admux/adx/bidder_version 比如包含版本信息等
- /admux/adx/bidder/bidder_xxx.json
- /admux/adx/bidder/bidder_yyy.json

content of /admux/adx/bidder_version:
202511012201

content of /admux/adx/bidder/bidder_xxx.json
{
    "version": "202511012201",
    "id": 1,
    "name": "xiaomi",
    "targeting": [  // boolean indexing expression
        {
            "segment": {"op": "in", "value": [1234]},
            "os": {"op": "=", "value": ["mobile"]}
        },
        {
        }
    ]

}

```

