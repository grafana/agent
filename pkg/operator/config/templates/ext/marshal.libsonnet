{
  // YAML marshals object to YAML.
  YAML(object):: std.native('marshalYAML')(object),

  // fromYAML unmarshals YAML text into an object.
  fromYAML(text):: std.native('unmarshalYAML')(text),

  // fromRawJSON unmarshals apiextensions.JSON into an object.
  // @param {apiext_v1.JSON} json
  fromRawJSON(json):: std.native('unmarshalRawJSON')(json),

  // intoStages unmarshals YAML text into []*PipelineStageSpec.
  // This is required because the "match" stage from Promtail is
  // recursive and you can't define recursive types in CRDs.
  intoStages(text):: std.native('intoStages')(text),
}
