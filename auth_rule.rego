package auth

import future.keywords

default allow := false

allow if {
    input.objectId == 0 # public objectId is 0
    actionFlag == 1 # access flag is 1
}

allow if {
    some role in input.userRoles
    input.objectId == role.objectId
    bits.and(input.actionFlag, role.actionFlags) != 0
}
