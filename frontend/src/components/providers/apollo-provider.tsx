'use client';

import * as React from 'react';
import { ApolloClient, InMemoryCache, ApolloProvider as Provider, createHttpLink } from '@apollo/client';
import { setContext } from '@apollo/client/link/context';
import { onError } from '@apollo/client/link/error';
import { toast } from 'react-hot-toast';

// HTTP Link for GraphQL requests
const httpLink = createHttpLink({
  uri: process.env.NEXT_PUBLIC_GRAPHQL_URL || 'http://localhost:8080/graphql',
});

// Auth link to add authentication headers
const authLink = setContext((_, { headers }) => {
  // Get token from localStorage or cookies
  const token = typeof window !== 'undefined' ? localStorage.getItem('auth_token') : null;
  const tenantId = typeof window !== 'undefined' ? localStorage.getItem('tenant_id') : null;
  
  return {
    headers: {
      ...headers,
      authorization: token ? `Bearer ${token}` : '',
      'x-tenant-id': tenantId || '',
    },
  };
});

// Error link for global error handling
const errorLink = onError(({ graphQLErrors, networkError, operation, forward }) => {
  if (graphQLErrors) {
    graphQLErrors.forEach(({ message, locations, path }) => {
      console.error(
        `[GraphQL error]: Message: ${message}, Location: ${locations}, Path: ${path}`
      );
      
      // Show user-friendly error messages
      if (message.includes('unauthorized') || message.includes('authentication')) {
        toast.error('Please log in to continue');
        // Redirect to login page if needed
        if (typeof window !== 'undefined' && !window.location.pathname.includes('/login')) {
          window.location.href = '/login';
        }
      } else if (message.includes('forbidden')) {
        toast.error('You do not have permission to perform this action');
      } else if (!message.includes('not found')) {
        // Don't show toast for "not found" errors as they're often expected
        toast.error('An error occurred. Please try again.');
      }
    });
  }

  if (networkError) {
    console.error(`[Network error]: ${networkError}`);
    
    if (networkError.message.includes('fetch')) {
      toast.error('Network error. Please check your connection.');
    } else if (networkError.message.includes('500')) {
      toast.error('Server error. Please try again later.');
    } else {
      toast.error('Connection error. Please try again.');
    }
  }
});

const client = new ApolloClient({
  link: errorLink.concat(authLink.concat(httpLink)),
  cache: new InMemoryCache({
    typePolicies: {
      Query: {
        fields: {
          // Add pagination support for lists
          users: {
            keyArgs: ['tenantId', 'filter'],
            merge(existing = [], incoming) {
              return [...existing, ...incoming];
            },
          },
          orders: {
            keyArgs: ['tenantId', 'filter'],
            merge(existing = [], incoming) {
              return [...existing, ...incoming];
            },
          },
          products: {
            keyArgs: ['tenantId', 'filter'],
            merge(existing = [], incoming) {
              return [...existing, ...incoming];
            },
          },
          reports: {
            keyArgs: ['tenantId', 'filter'],
            merge(existing = [], incoming) {
              return [...existing, ...incoming];
            },
          },
        },
      },
    },
  }),
  defaultOptions: {
    watchQuery: {
      errorPolicy: 'all',
    },
    query: {
      errorPolicy: 'all',
    },
    mutate: {
      errorPolicy: 'all',
    },
  },
});

interface ApolloProviderProps {
  children: React.ReactNode;
}

export function ApolloProvider({ children }: ApolloProviderProps) {
  return <Provider client={client}>{children}</Provider>;
}
