export interface Resource {
  group: string;
  kind: string;
  version: string;
  namespaced: boolean;
  list: boolean;
  watch: boolean;
}
