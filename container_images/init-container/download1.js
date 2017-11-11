/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenX Cloud. All Rights Reserved.
 * v0.1 - 2017-06-30
 * @author Zhangpc
 *
 */

/**
 * download ci scripts
 */

'use strict'

const crypto = require('crypto')
const fs = require('fs')
const urllib = require('urllib')
const SECRET_KEY = 'dazyunsecretkeysforuserstenx20141019generatedKey'
const ENV = process.env
const SCRIPT_ENTRY_INFO = ENV.SCRIPT_ENTRY_INFO
const SCRIPT_URL = ENV.SCRIPT_URL
// const SCRIPT_ENTRY_INFO = 'K9/N3ccZiEr85Y/zMKLjUGob3HvgWbV+9oM0Hv/2LAeebeQ8+KVQZ8nvlk9VlrP0AuSYF3ex93O8b+qI+kddssdHEgdi/I6pgf7ZZGlXw6R5wu99Z1+Vk0yII//t3nQhrTit27uY7uG08eCezvIBMlah9gLFRph36KU8mDlB4O69zBATmQP6AudzRgh6HCm9zhsmjBr0ImC5fRummhpNFJyXHTYm'
// const SCRIPT_URL = 'https://paasdev.enncloud.cn:38090/api/v2/devops/ci-scripts'

// Disabled rejecting self-signed certificates
process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0'

if (!SCRIPT_ENTRY_INFO) {
    console.error("SCRIPT_ENTRY_INFO is empty,please retry ")
    process.exit(0)
    return
}

const scriptEntryInfo = aeadDecrypt(SCRIPT_ENTRY_INFO).split(':')
const scriptId = scriptEntryInfo[0]
const username = scriptEntryInfo[1]
const token = scriptEntryInfo[2]
const scriptPath = `/app/${scriptId}`

const scriptUrl = `${SCRIPT_URL}/${scriptId}`
const reqOptions = {
    dataType: 'json',
    contentType: 'json',
    timeout: 1000 * 60,
    headers: {
        username,
        authorization: `token ${token}`,
    }
}

fetchForeverUnitlSuccess(1).then(res => {
    const script = res.data.script
    fs.writeFileSync(scriptPath, script.content, { mode: 0o755 })
if (!script.content){
    console.error("script content is empty please retry again")
    process.exit(-1)

}else{
    console.log("script content is ==>:%s"+JSON.stringify(script.content))
    process.exit(0)

}
}).catch(err => {
    console.error(err.stack);
process.exit(-1)
})

function fetchForeverUnitlSuccess(num) {

    if (num === 6) {
        return Promise.reject(`超过重试次数，下载脚本失败`);
    }

    console.log("==============>> request "+num+" times")

    return urllib.request(scriptUrl, reqOptions).then(res => {
        if (res.status !== 200) {
        console.error(res);
        throw new Error(JSON.stringify(res))
    }
    return res
}).catch(err => {

        console.log(err.stack);
    return fetchForeverUnitlSuccess(++num);

})
}

function aeadDecrypt(encrypted) {
    const buffer = new Buffer(encrypted, 'base64')
    const salt = buffer.slice(0, 64)
    const iv = buffer.slice(64, 76)
    const tag = buffer.slice(76, 92)
    const content = buffer.slice(92)
    const secret = new Buffer(SECRET_KEY)
    const key = crypto.pbkdf2Sync(secret, salt, 2145, 32, 'sha512')
    const decipher = crypto.createDecipheriv('aes-256-gcm', key, iv)
    decipher.setAuthTag(tag)
    return decipher.update(content, 'binary', 'utf8') + decipher.final('utf8')
}