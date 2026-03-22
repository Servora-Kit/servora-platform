import { createRequestHandler } from '@servora/web-pkg/request'
import type { RequestHandlerOptions } from '@servora/web-pkg/request'

import {
  createApplicationServiceClient,
  createAuthnServiceClient,
  createUserServiceClient,
} from '@servora/api-client/servora/iam/service/v1/index'
import type {
  ApplicationService,
  AuthnService,
  UserService,
} from '@servora/api-client/servora/iam/service/v1/index'

export interface IamClients {
  authn: AuthnService
  user: UserService
  application: ApplicationService
}

export function createIamClients(
  options: RequestHandlerOptions = {},
): IamClients {
  const handler = createRequestHandler(options)

  return {
    authn: createAuthnServiceClient(handler),
    user: createUserServiceClient(handler),
    application: createApplicationServiceClient(handler),
  }
}

export type { RequestHandlerOptions } from '@servora/web-pkg/request'
export { ApiError } from '@servora/web-pkg/request'
export type { ApiErrorKind, TokenStore, RequestHandler } from '@servora/web-pkg/request'
