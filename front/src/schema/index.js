import * as yup from 'yup'

export const SchemaOutGwTrans = yup.object({
  type: yup.string().label('类型').required(),
  param: yup
    .array()
    .when('type', {
      is: 'prefix',
      then: () => yup.tuple([yup.string().label('前缀').required().trim()])
    })
    .when('type', {
      is: 'suffix',
      then: () => yup.tuple([yup.string().label('后缀').required().trim()])
    })
    .when('type', {
      is: 'replace',
      then: () =>
        yup.tuple([
          yup.string().label('正则').required().trim(),
          yup.string().label('替换内容').trim()
        ])
    })
})

export const SchemaOutGw = yup.object({
  name: yup.string().label('网关名称').trim().required(),
  protocol: yup.string().label('协议').required(),
  addr: yup.string().label('地址').trim().required(),
  enable: yup.boolean().label('启用').required(),
  options: yup.object({
    password: yup.string().label('密码'),
    registry: yup.boolean().label('是否登陆')
  }),
  transcaller: yup.array().of(SchemaOutGwTrans),
  transcallee: yup.array().of(SchemaOutGwTrans)
})

export const SchemaNumber = yup.object({
  number: yup.string().label('号码').required(),
  outgw: yup.string().label('网关').required(),
  enable: yup.boolean().label('启用').required(),
  tag: yup.object({
    province: yup.string().label('省'),
    city: yup.string().label('市')
  })
})

export const SchemaObjective = yup.object({
  title: yup.string().label('目标名称').trim().required().min(10, '目标名称长度不能少于10个字符'),
  info: yup.object({
    company: yup.string().label('公司名称').trim().required(),
    background: yup.string().label('背景信息').trim().required()
  })
})

export const SchemaConfigDial = yup.object({
  caller: yup.object({
    affinity: yup.boolean().label('亲和性呼叫')
  })
})

export const SchemaConfigPrivacy = yup.object({
  hideNumber: yup.boolean().label('隐藏号码')
})

export const SchemaConfigCloud = yup.object({
  addr: yup.string().label('服务器地址').trim(),
  appid: yup.string().label('应用ID').trim(),
  secret: yup.string().label('密钥').trim(),
  lifecycle: yup.object({
    beforeCall: yup.object({
      blacklist: yup.boolean().label('黑名单检查'),
      flashCard: yup.boolean().label('闪信通知')
    })
  })
})

export const SchemaConfigIce = yup.array().of(
  yup
    .object({
      urls: yup.string().label('服务器地址').required(),
      username: yup.string().label('用户名'),
      redential: yup.string().label('密码')
    })
    .noUnknown()
)

export const SchemaUser = yup.object({
  name: yup.string().label('姓名').trim().required(),
  email: yup.string().label('邮箱').email().required().trim(),
  password: yup.string().label('密码').min(6),
  isAdmin: yup.boolean().label('管理员'),
  verified: yup.boolean().label('已验证')
})

export const SchemaTask = yup.object({
  desc: yup.string().label('任务描述').required(),
  contact: yup.string().label('客户名称').required(),
  callee: yup.string().label('联系方式').required(),
  own: yup.string().label('负责人').required()
})
