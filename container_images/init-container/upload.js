/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * v0.1 - 2016-11-23
 * @author Huang Xin
 * 
 */

/**
 * upload Dockerfile and Readme
 */
'use strict'

const urllib = require('urllib')
const fs = require('fs')

const UPLOAD_URL = process.env.UPLOAD_URL
const AUTH = process.env.AUTH ? process.env.AUTH : ""
const IMAGE_NAME = process.env.IMAGE_NAME
const FILES_PATH = process.env.FILES_PATH
const CONTRIBUTOR = process.env.CONTRIBUTOR
const DOCKERFILE_NAME = process.env.DOCKERFILE_NAME ? process.env.DOCKERFILE_NAME : "Dockerfile"
const AdminUserName = 'admin'

function _createRequest(url, options, callback) {
  return urllib.request(url, options, callback)
}

function _getImageInfo(callback) {
  const requestUrl = `${UPLOAD_URL}/image/${IMAGE_NAME}`
  const options = {
    dataType: 'json',
    headers: _getAuthorizationHeader()
  }
  return _createRequest(requestUrl, options, callback)
}

function _updateImageInfo(image, callback) {
  const requestUrl = `${UPLOAD_URL}/image/${IMAGE_NAME}`
  const options = {
    method: 'POST',
    dataType: 'json',
    data: image,
    headers: _getAuthorizationHeader()
  }
  return _createRequest(requestUrl, options, callback)
}

function  _getAuthorizationHeader() {
  let authHeader = {
    'Authorization': AUTH
  }
  var authUser = ''
  var auth = AUTH.split(' ');
  // logger.info({headers: req.headers, url: req.url});
  if (auth[0] == 'Basic') {
    var buff  = new Buffer(auth[1], 'base64');
    var plain = buff.toString();
    var creds = plain.split(':');
    authUser = creds[0];
  }
  // Only admin user can use onbehalfUser
  if (authUser == AdminUserName && CONTRIBUTOR) {
    authHeader.onbehalfuser = CONTRIBUTOR;
  }
  return authHeader
}

function _getFile(path, callback) {
  fs.exists(path, function (exists) {
    if (!exists) {
      console.warn(`There is no ${path}`)
      return callback(null, "")
    } 
    fs.readFile(path, function (err, data) {
      callback(err, data.toString())
    })
  })
}

if (!UPLOAD_URL) {
  console.error('UPLOAD_URL should be specified')
  return
}

if (!IMAGE_NAME) {
  console.error('IMAGE_NAME should be specified')
  return
}

if (!FILES_PATH) {
  console.error('FILES_PATH should be specified')
  return
}
console.info(`Reading files from ${FILES_PATH}...`)

_getFile(FILES_PATH + '/README.md', function (err, readme) {
  if (err) {
    console.warn('Failed to read README:', err)
    return 
  }
  _getFile(FILES_PATH + `/${DOCKERFILE_NAME}`, function (err, dockerfile) {
    if (err) {
      console.warn(`Failed to read ${DOCKERFILE_NAME}:`, err)
      return
    }
    if (!dockerfile && !readme) {
      return
    } else {
      console.log("File uploaded")
    }
    _getImageInfo(function (err, data, resp) {
      if (err) {
        console.warn('Failed to get image info:', data || err.name)
        return
      }
      if (resp.statusCode > 299 && resp.statusCode !== 404) {
        console.warn(`Failed to get image info: ${resp.statusCode}-${data}`)
        return
      }
      let updated = false
      let info = {}
      if (resp.statusCode === 404) {
        updated = true
        info.dockerfile = dockerfile
        info.detail = readme
        //使用admin创建，这样用户无法看到镜像。待push之后再修改contributor为用户
        info.contributor = 'admin'
        // Default will be private
        info.isPrivate = 1
      } else {
        if (data.dockerfile !== dockerfile) {
          updated = true
          info.dockerfile = dockerfile
        }
        if (data.detail !== readme) {
          updated = true
          info.detail = readme
        }
        // If not isPrivate defined, make it as private
        if (!data.isPrivate) {
          info.isPrivate = 1
        }
      }
      if (updated) {
        _updateImageInfo(info, function (err, data, resp) {
          if (err) {
            console.warn('Failed to update image info:', data || err.name)
            return
          }
          if (resp.statusCode > 299) {
            console.warn(`Failed to update image info: ${resp.statusCode}-${data}`)
            return
          }
          console.info('Upload successfully')
        })
        return 
      }
      console.info('No files need to be updated')
    })
  })
})