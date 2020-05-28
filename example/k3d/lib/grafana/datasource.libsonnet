{
  new(name, url, default=false, method='GET', type='prometheus'):: {
    apiVersion: 1,
    datasources: [{
      name: name,
      type: type,
      access: 'proxy',
      url: url,
      isDefault: default,
      version: 1,
      editable: false,
      jsonData: {
        httpMethod: method,
      },
    }],
  },

  withBasicAuth(username, password):: {
    datasources: std.map(function(ds) ds {
      basicAuth: true,
      basicAuthUser: username,
      basicAuthPassword: password,
    }, super.datasources),
  },
}
