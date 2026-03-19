import { createRequestHandler } from './requestHandler'
import type { RequestHandlerOptions } from './requestHandler'

import {
  createApplicationServiceClient,
  createAuthnServiceClient,
  createUserServiceClient,
} from '@servora/api-client/iam/service/v1/index'
import type {
  ApplicationService,
  AuthnService,
  UserService,
} from '@servora/api-client/iam/service/v1/index'

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

export type { RequestHandlerOptions } from './requestHandler'
export { ApiError } from './requestHandler'
export type { ApiErrorKind, TokenStore, RequestHandler } from './requestHandler'
