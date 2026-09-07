.server.name == $expected[0].name and
.server.version == $expected[0].version and
.server.packages == $expected[0].packages and
.server.repository == $expected[0].repository and
._meta["io.modelcontextprotocol.registry/official"].status == "active"
