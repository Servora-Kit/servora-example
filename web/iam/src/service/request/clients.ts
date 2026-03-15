import {
  createRequestHandler,
  type RequestHandlerOptions,
} from './requestHandler'

import {
  type ApplicationService,
  createApplicationServiceClient,
  type AuthnService,
  createAuthnServiceClient,
  type OrganizationService,
  createOrganizationServiceClient,
  type ProjectService,
  createProjectServiceClient,
  type UserService,
  createUserServiceClient,
} from '#/service/gen/iam/service/v1/index'

export interface IamClients {
  authn: AuthnService
  user: UserService
  organization: OrganizationService
  project: ProjectService
  application: ApplicationService
}

export function createIamClients(
  options: RequestHandlerOptions = {},
): IamClients {
  const handler = createRequestHandler(options)

  return {
    authn: createAuthnServiceClient(handler),
    user: createUserServiceClient(handler),
    organization: createOrganizationServiceClient(handler),
    project: createProjectServiceClient(handler),
    application: createApplicationServiceClient(handler),
  }
}

export type { RequestHandlerOptions } from './requestHandler'
export { ApiError } from './requestHandler'
export type { ApiErrorKind, TokenStore, RequestHandler } from './requestHandler'
