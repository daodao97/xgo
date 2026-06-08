function apiBaseURL() {
  return (window.BASE_URL || window.OmsOptions?.axios?.baseURL || '/api').replace(/\/$/, '')
}

function loginForm() {
  return document.querySelector('#login #form')
}

function usernameInput(form) {
  return form?.querySelector('input[placeholder="用户名"]')
}

function emailCodeInput(form) {
  return form?.querySelector('.input-email-code')
}

let emailCodeEnabled = false
let emailCodeStatusLoaded = false

async function loadEmailCodeEnabled() {
  if (emailCodeStatusLoaded) return emailCodeEnabled

  emailCodeStatusLoaded = true
  try {
    const resp = await window.fetch(`${apiBaseURL()}/user/email/code/status`, {
      method: 'GET',
      credentials: 'include'
    })
    const data = await resp.json()
    emailCodeEnabled = data.code === 0 && data.data?.enabled === true
  } catch (_) {
    emailCodeEnabled = false
  }
  return emailCodeEnabled
}

function message(text, type = 'info') {
  if (window.ElMessage) {
    window.ElMessage({ message: text, type })
    return
  }
  const form = loginForm()
  if (!form) return
  let tips = form.querySelector('.email-code-tips')
  if (!tips) {
    tips = document.createElement('div')
    tips.className = 'email-code-tips'
    form.appendChild(tips)
  }
  tips.textContent = text
}

function startCountdown(button, seconds = 60) {
  let remain = seconds
  button.disabled = true
  button.textContent = `${remain}s`
  const timer = window.setInterval(() => {
    remain -= 1
    button.textContent = remain > 0 ? `${remain}s` : '发送验证码'
    if (remain <= 0) {
      button.disabled = false
      window.clearInterval(timer)
    }
  }, 1000)
}

async function sendEmailCode(button) {
  const form = loginForm()
  const username = usernameInput(form)?.value.trim()
  if (!username) {
    message('请先输入用户名', 'error')
    return
  }

  button.disabled = true
  button.textContent = '发送中'
  try {
    const resp = await window.fetch(`${apiBaseURL()}/user/email/code`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ username })
    })
    const data = await resp.json()
    if (data.code !== 0) {
      throw new Error(data.message || data.msg || '验证码发送失败')
    }
    message('验证码已发送', 'success')
    startCountdown(button)
  } catch (err) {
    message(err?.message || '验证码发送失败', 'error')
    button.disabled = false
    button.textContent = '发送验证码'
  }
}

function ensureEmailCodeField() {
  const form = loginForm()
  if (!form) return

  if (!emailCodeEnabled) {
    form.querySelector('.email-code')?.remove()
    return
  }
  if (form.querySelector('.email-code')) return

  const row = document.createElement('div')
  row.className = 'email-code'

  const input = document.createElement('input')
  input.type = 'text'
  input.placeholder = '邮箱验证码'
  input.className = 'input-email-code'

  const button = document.createElement('button')
  button.type = 'button'
  button.className = 'email-code-button el-button el-button--primary'
  button.textContent = '发送验证码'
  button.addEventListener('click', () => sendEmailCode(button))

  row.append(input, button)
  const loginButton = form.querySelector('.login-button')
  form.insertBefore(row, loginButton || form.firstChild)
}

function patchLoginPayload() {
  if (window.__xgoLoginEmailCodePatched) return
  window.__xgoLoginEmailCodePatched = true

  const rawOpen = XMLHttpRequest.prototype.open
  const rawSend = XMLHttpRequest.prototype.send

  XMLHttpRequest.prototype.open = function (method, url, ...rest) {
    this.__xgoMethod = method
    this.__xgoURL = String(url || '')
    return rawOpen.call(this, method, url, ...rest)
  }

  XMLHttpRequest.prototype.send = function (body) {
    const isLogin = String(this.__xgoMethod || '').toUpperCase() === 'POST' && this.__xgoURL.includes('/user/login')
    const code = emailCodeInput(loginForm())?.value.trim()
    if (isLogin && code && typeof body === 'string') {
      try {
        const data = JSON.parse(body)
        if (!data.code && !data.email_code) {
          data.code = code
          body = JSON.stringify(data)
        }
      } catch (_) {}
    }
    return rawSend.call(this, body)
  }
}

export function setupLoginEmailCode() {
  patchLoginPayload()
  const ensure = async () => {
    await loadEmailCodeEnabled()
    window.requestAnimationFrame(ensureEmailCodeField)
  }
  window.addEventListener('hashchange', ensure)
  new MutationObserver(ensure).observe(document.body, { childList: true, subtree: true })
  ensure()
}
