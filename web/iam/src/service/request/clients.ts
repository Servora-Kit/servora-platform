import { createRequestHandler } from './requestHandler'
import type { RequestHandlerOptions } from './requestHandler'

import {
  createApplicationServiceClient,
  createAuthnServiceClient,
  createOrganizationServiceClient,
  createTenantServiceClient,
  createUserServiceClient,
} from '@servora/api-client/iam/service/v1/index'
import type {
  ApplicationService,
  AuthnService,
  OrganizationService,
  TenantService,
  UserService,
} from '@servora/api-client/iam/service/v1/index'

export interface IamClients {
  authn: AuthnService
  user: UserService
  organization: OrganizationService
  application: ApplicationService
  tenant: TenantService
}

export function createIamClients(
  options: RequestHandlerOptions = {},
): IamClients {
  const handler = createRequestHandler(options)

  return {
    authn: createAuthnServiceClient(handler),
    user: createUserServiceClient(handler),
    organization: createOrganizationServiceClient(handler),
    application: createApplicationServiceClient(handler),
    tenant: createTenantServiceClient(handler),
  }
}

export type { RequestHandlerOptions } from './requestHandler'
export { ApiError } from './requestHandler'
export type { ApiErrorKind, TokenStore, RequestHandler } from './requestHandler'
