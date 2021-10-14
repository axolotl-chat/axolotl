const uuidV4Regex = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
const isValidV4UUID = uuid => uuidV4Regex.test(uuid);
export const validateUUID = function (uuid) {
  var result = isValidV4UUID(uuid)
  return result
}
