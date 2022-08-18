import React from 'react';

/**
 * PathPrefixContext propgates the base URL throughout the component tree where
 * the application is hosted.
 */
const PathPrefixContext = React.createContext('');

/**
 * usePathPrefix retrieves the current base URL where the application is
 * hosted. Links and API calls should be all relative to this path. Returns
 * `/` if there is no path prefix.
 */
function usePathPrefix(): string {
  const prefix = React.useContext(PathPrefixContext);
  if (prefix === '') {
    return '/';
  }
  return prefix;
}

export { PathPrefixContext, usePathPrefix };
