Scope noder er atomiske og indeholder en valgfri primary key navn (scope navnet man requester)
Beskrivelse og titel ligger på publish regel fra RS -> Scope

Hvis en client requester et scope uden audience, betyder det at man spørger om adgang til alle resource servere som publisher dette scope.
Hvis man angiver audiences requester man kun til disse specifikke RS's

Man skal kunne give starten af may grant chain videre til en anden.
Dette vil gøre at en anden får starten af chain og dermed også kan give denne videre
Det svare til at få root for et scope til en resource server
