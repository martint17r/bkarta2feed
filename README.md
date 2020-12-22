

# Laufende Fusionskontrollverfahren als Feed

Dieses Skriptchen parst die Tabelle auf
https://www.bundeskartellamt.de/DE/Fusionskontrolle/LaufendeVerfahren/laufendeverfahren_node.html
und gibt einen Atom Feed auf Standard Out aus.

```bash
./bkarta2feed https://www.bundeskartellamt.de/DE/Fusionskontrolle/LaufendeVerfahren/laufendeverfahren_node.html
```

## Todos

* Beendete oder abgebrochene Verfahren werden nicht gesondert berücksichtigt
* Die Branche und das optionale Bundesland werden an die Description angehängt
