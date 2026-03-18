import { useMatches, Link } from '@tanstack/react-router'
import { useQueries } from '@tanstack/react-query'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '#/components/ui/breadcrumb'
import { Fragment } from 'react'
import { iamClients } from '#/api'

interface BreadcrumbSegment {
  label: string
  href: string
}

const ROUTE_LABELS: Record<string, string> = {
  dashboard: '概览',
  organizations: '组织',
  applications: '应用',
  users: '用户',
  tenants: '租户',
  settings: '设置',
  members: '成员管理',
  profile: '个人信息',
  security: '安全',
}

interface DynamicSegment {
  placeholder: string
  paramName: string
  paramValue: string
}

function getEntityQueryConfig(paramName: string, paramValue: string) {
  switch (paramName) {
    case 'orgId':
      return {
        queryKey: ['organization', paramValue] as const,
        queryFn: () => iamClients.organization.GetOrganization({ id: paramValue }),
      }
    case 'userId':
      return {
        queryKey: ['user', paramValue] as const,
        queryFn: () => iamClients.user.GetUser({ id: paramValue }),
      }
    case 'tenantId':
      return {
        queryKey: ['tenant', paramValue] as const,
        queryFn: () => iamClients.tenant.GetTenant({ id: paramValue }),
      }
    case 'appId':
      return {
        queryKey: ['application', paramValue] as const,
        queryFn: () => iamClients.application.GetApplication({ id: paramValue }),
      }
    default:
      return null
  }
}

function extractEntityName(paramName: string, data: unknown): string | undefined {
  if (!data) return undefined
  switch (paramName) {
    case 'orgId': {
      const d = data as { organization?: { displayName?: string; name?: string } }
      return d.organization?.displayName || d.organization?.name || undefined
    }
    case 'userId': {
      const d = data as { user?: { name?: string } }
      return d.user?.name || undefined
    }
    case 'tenantId': {
      const d = data as { tenant?: { displayName?: string; name?: string } }
      return d.tenant?.displayName || d.tenant?.name || undefined
    }
    case 'appId': {
      const d = data as { application?: { name?: string } }
      return d.application?.name || undefined
    }
    default:
      return undefined
  }
}

// Parse the leaf routeId into logical path segments, stripping layout prefixes (_app, _platform, etc.)
function parseLogicalParts(routeId: string): string[] {
  return routeId
    .split('/')
    .filter(Boolean)
    .filter((p) => !p.startsWith('_'))
}

export function AppBreadcrumb() {
  const matches = useMatches()

  const leafMatch = matches.length > 0 ? matches[matches.length - 1] : null
  const params = (leafMatch?.params ?? {}) as Record<string, string>
  const logicalParts = leafMatch ? parseLogicalParts(leafMatch.routeId) : []

  const dynamicSegments: DynamicSegment[] = logicalParts
    .filter((p) => p.startsWith('$'))
    .map((p) => ({
      placeholder: p,
      paramName: p.slice(1),
      paramValue: params[p.slice(1)] ?? '',
    }))

  // Fetch entity display names via React Query — deduped with queries already made by the page.
  const entityQueries = useQueries({
    queries: dynamicSegments.map(({ paramName, paramValue }) => {
      const config = getEntityQueryConfig(paramName, paramValue)
      if (!config || !paramValue) {
        return {
          queryKey: ['_noop', paramName, paramValue] as const,
          queryFn: (): null => null,
          enabled: false,
        }
      }
      return { ...config, enabled: true, staleTime: 60_000 }
    }),
  })

  // Map each dynamic placeholder to its resolved display name.
  const entityNameMap = new Map<string, string>()
  dynamicSegments.forEach(({ placeholder, paramName, paramValue }, idx) => {
    const name = extractEntityName(paramName, entityQueries[idx]?.data)
    // Fallback to first 8 chars of the raw ID while loading.
    entityNameMap.set(placeholder, name ?? paramValue.slice(0, 8))
  })

  // Build breadcrumb segments, accumulating the href as we go.
  const segments: BreadcrumbSegment[] = []
  let accumulatedPath = ''

  for (const part of logicalParts) {
    const isDynamic = part.startsWith('$')
    const actualValue = isDynamic ? (params[part.slice(1)] ?? part) : part
    accumulatedPath += '/' + actualValue

    const label = isDynamic
      ? (entityNameMap.get(part) ?? actualValue.slice(0, 8))
      : (ROUTE_LABELS[part] ?? part)

    segments.push({ label, href: accumulatedPath })
  }

  if (segments.length === 0) return null

  return (
    <Breadcrumb>
      <BreadcrumbList>
        {segments.map((seg, idx) => (
          <Fragment key={seg.href}>
            {idx > 0 && <BreadcrumbSeparator />}
            <BreadcrumbItem>
              {idx < segments.length - 1 ? (
                <BreadcrumbLink asChild>
                  <Link to={seg.href as '/'}>{seg.label}</Link>
                </BreadcrumbLink>
              ) : (
                <BreadcrumbPage>{seg.label}</BreadcrumbPage>
              )}
            </BreadcrumbItem>
          </Fragment>
        ))}
      </BreadcrumbList>
    </Breadcrumb>
  )
}
