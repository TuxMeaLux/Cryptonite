# Contest for surely not generic malware

## Abstract
This malware is intended to be a backdoor for the commonest OSs. It should be written in GO to let make a static binary for a stand alone deployment.

It should be composed of two parts: *stager* and *backdoor*.
The first one is the one that will be injected and executed into the victim machine, has task to download and deploy the right architecture backdoor from a specific server. The second one is the real malware and has to implement the functions listed below.

## Target OSs
 1. Windows 7/8/8.1/10
 2. GNU/Linux
 3. Mac OSX

## Functions
 * [Certificate pinning with double couple keys] (#certificate-pinning-with-double-couple-keys-notes)
 * Proxy (ssl tunneling?)
 * POST and GET for files and directories (sshfs)
 * Keylogger
 * Grab saved password in the browser
 * Grab LUKS master key
 * Reverse shell
 * Take screenshot
 * Connection to C2s (Command and Control Server)
  > C2s as hidden service (connect bots through tor)

 * Disinstallation

Each functions should work on each target OS.

### Certificate pinning with double couple keys _(notes)_
La connessione TLS viene iniziata dal client. Il client al suo interno ha hardcodato il certificato della nostra CA ed accetterà di connettersi ad un server solo se questo gli fornisce un certificato valido, ovvero firmato dalla nostra CA.

In questo modo possiamo assicurarsi che i clients si connetteranno solamente con i server da noi autorizzati.

Ovviamente la chiave privata del server dovrà essere deployata insieme al server stesso, ed una violazione di quest'ultimo non è da escludere, questo potrebbe comportare che qualcuno si **impossessi** della chiave privata del nostro C2s...

Questo qualcuno quindi potrebbe impartire comandi ai clients. Ed è qui che entra in gioco la seconda coppia di chiavi per tutelarci.

#### Doble couple keys
Un'altra coppia di chiavi viene usata per firmare ogni comando che viene impartito, e questo viene fatto direttamente sulla macchina del botmaster, in modo che la chiave privata che firma i messaggi non debba essere caricata in rete, ma può rimanere isolata (offline) nella macchina del botmaster.

I client avranno quindi hardcodata la chiave pubblica di questa seconda coppia di chiavi, che utilizzeranno per validare la firma per ogni messaggio che gli arriva.

Ogni volta che un comando viene impartito dovrà avere una firma diversa da tutti i comandi inviati precedentemente altrimenti, se intercettato, potrebbe essere riusato all'infinto.

Per fare questo verrà generato un nonce dal botmaster ogni volta che un comando venga impartito, sarà fatto sia se si vuole comandare un solo client o un insieme di questi. Una volta generato viene comunicato a al/ai client interessati, che lo memorizzeranno.

Successivamente il botmaster appende al comando che vuole impartire il nonce, in una forma del tipo: comando+"\n"+nonce (dovrà essere abbastanza lungo >=128 chars).

Così facendo questo comando firmato sarà valido solo per una voltà, ovvero un **on-time-command**... in quanto se si volesse rimpartire lo stesso comando una seconda volta, bisognerebbe prima generare un nuovo nonce, ricondividerlo e quindi generare una firma valida.

I client non accettano che lo stesso nonce sia utilizzato per impartire due comandi consecutivi.

### Certificate Revocation

Grazie al sistema della doppia coppia di chiavi, se venisse rubata la chiave privata del server ci potremmo comunque collegare ai clients in quanto il certificato sarebbe ancora funzionante (se chi ha violato il server non lo abbia cancellato, **Per questo è importante avere un backup del certificato e della chiave privata del server in produzione**), e impartire come ordine di inserire quella chiave nella lista dei certificati revocati (anchessi hardcodati nella backdoor). Quindi creare un nuovo certificato per il nuovo server, firmarlo con la nostra CA ed eventualmente impartire l'ordine ai clients di connetteri ad un nuovo indirizzo ip.

### Socket Client/Server
 * Far comunicare client e server via socket

### Steps with RSA
 1. Generare chiavi rsa.GenerateKey(rand.Reader, 2048)
 2. Codificarle in PEM per salvarle

```
import (
    "crypto/x509"
    "encoding/pem"
  )

x509.MarshalPKCS1PrivateKey(key *rsa.PrivateKey) converts a private key to ASN.1 DER encoded form.
type Block struct
    Bytes   []byte // The decoded bytes of the contents. Typically a DER encoded ASN.1 structure.
pem.Encode(out io.Writer, b *Block)
```
 3. Dare la pubkey al client e la prvkey al server
 4. Server: scrive un messaggio, lo firma e lo invia al client
 5. Client: controlla la firma, se è valida stampa il messaggio
