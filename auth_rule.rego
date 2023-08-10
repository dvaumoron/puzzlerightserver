package auth

import future.keywords

default allow := false

publicObject := input.objectId == 0

allow if {
    publicObject
    input.actionFlag == 1 # access flag is 1
}

allow if {
    publicObject
    input.userId != 0 # connected user
}

allow if {
    some role in input.userRoles
    input.objectId == role.objectId
    bits.and(input.actionFlag, role.actionFlags) != 0
}
