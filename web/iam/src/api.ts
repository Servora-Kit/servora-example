import { env } from '#/env'
import { createIamClients, ApiError, type TokenStore } from '#/service/request/clients'

const ACCESS_TOKEN_KEY = 'iam_access_token'
const REFRESH_TOKEN_KEY = 'iam_refresh_token'

export const tokenStore: TokenStore = {
  getAccessToken() {
    return localStorage.getItem(ACCESS_TOKEN_KEY)
  },
  getRefreshToken() {
    return localStorage.getItem(REFRESH_TOKEN_KEY)
  },
  setTokens(accessToken: string, refreshToken: string) {
    localStorage.setItem(ACCESS_TOKEN_KEY, accessToken)
    localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken)
  },
  clear() {
    localStorage.removeItem(ACCESS_TOKEN_KEY)
    localStorage.removeItem(REFRESH_TOKEN_KEY)
  },
}

export const iamClients = createIamClients({
  baseUrl: env.VITE_API_BASE_URL ?? '',
  tokenStore,
  timeoutMs: 30_000,
  autoRefreshToken: true,
  onError(error: ApiError) {
    if (error.httpStatus === 401) {
      tokenStore.clear()
      const currentPath = window.location.pathname
      window.location.href = `/login?redirect=${encodeURIComponent(currentPath)}`
    }
  },
})
