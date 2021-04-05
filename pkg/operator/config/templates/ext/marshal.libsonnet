{
  // YAML marshals object to YAML.
  YAML(object):: std.native('marshalYAML')(object),

  // fromYAML unmarshals YAML text into an object.
  fromYAML(text):: std.native('unmarshalYAML')(text),
}
