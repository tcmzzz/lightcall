const helper = (obj, path, value) => {
  let [current, ...rest] = path

  if (rest.length > 0) {
    if (!obj[current]) {
      const isNumber = `${+rest[0]}` === rest[0]
      obj[current] = isNumber ? [] : {}
    }

    if (typeof obj[current] !== 'object') {
      const isNumber = `${+rest[0]}` === rest[0]
      obj[current] = helper(isNumber ? [] : {}, rest, value)
    } else {
      obj[current] = helper(obj[current], rest, value)
    }
  } else {
    obj[current] = value
  }

  return obj
}

export const SetNestObj = (obj, path, value) => {
  let pathArr = path

  if (typeof path === 'string') {
    pathArr = path.replaceAll('[', '.').replaceAll(']', '').split('.')
  }

  helper(obj, pathArr, value)
}

export const GenValidFn = (sche, objVm, errVm) => {
  return async (path) => {
    SetNestObj(errVm.value, path, null)

    let obj = JSON.parse(JSON.stringify(objVm.value))
    try {
      obj = sche.cast(obj)
    } catch (e) {}

    try {
      await sche.validateAt(path, obj)
    } catch (e) {
      if (e.path) {
        SetNestObj(errVm.value, e.path, e.message)
      } else {
        console.error('unexpected error:', e.message)
      }
    }
  }
}

export const GenSaveFn = (sche, objVm, errVm, saveFn) => {
  return async () => {
    let obj = JSON.parse(JSON.stringify(objVm.value))
    try {
      obj = sche.cast(obj)
    } catch (e) {
      console.error('cast fail', e.toString())
    }

    try {
      await sche.validate(obj, { abortEarly: false })
      saveFn(obj)
    } catch (e) {
      const obj = {}
      e.inner.forEach((err) => {
        SetNestObj(obj, err.path, err.message)
      })
      errVm.value = obj
    }
  }
}
