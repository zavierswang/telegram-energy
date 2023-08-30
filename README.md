# 能量租赁机器人

![group-title.png](https://github.com/zavierswang/telegram-energy/blob/main/assets/group-title.png)

### 功能
* 🔥 使用全新租用模式，更低的价格租用能量，最高可节省87%
* 🔥 更新了默认能量数为一次，避免多数用户浪费能量
* 🔥 优化交互逻辑，用户支付地址即为租赁能量接收地址


### 更多程序
* [telegram-trx](https://github.com/zavierswang/telegram-trx) **TRX兑换机器人**
* [telegram-monitor](https://github.com/zavierswang/telegram-monitor) **TRC20钱包事件监听机器人**
* [telegram-search](https://github.com/zavierswang/telegram-search) **导航机器人**（可支持全网搜索，API收费有点小贵）
* [telegram-premium](https://github.com/zavierswang/telegram-premium) **Telegram Premium自动充值机器人**
* [telegram-replay](https://github.com/zavierswang/telegram-replay) **双向机器人**
* [telegram-energy](https://github.com/zavierswang/telegram-energy) **TRON能量租凭机器人**
* [telegram-proto](https://github.com/zavierswang/telegram-proto) **Telegram协议号机器人**


### 部署
* 本程序基于Telegram Bot，分为主程序telegram-energy
* 确保机器人部署服务可以访问外网`telegram.org`
* 使用自己的telegram生成一个机器人，并获取到token
* 更换`tron_scan_key`和`tron_grid_key`
* 配置文件`telegram-energy.yaml.example`改名为`telegram-energy.yaml`

> **注意：**
> * 不支持交易所转帐事件监听
> * 对linux不熟悉的给点打赏手把手教学🤭
> * 配置文件中的`license`配置请找 [🫣我](https://t.me/tg_llama) 拿~