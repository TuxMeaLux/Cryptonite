# Cryptonite

## Abstract
This malware is intended to be a backdoor for the commonest OSes. It's been written in GO to easy cross-compiling static binary for different architectures and OSes. That also simplify future deployments.

It's composed of two parts: *stager* and *backdoor*.
 * The first one is the one that will be injected and executed into the victim machine exploiting a vector. The main task of this one is to understand the environment, what kind of OSes and architecture it's running on, then download and deploy the right architecture backdoor from a specific server.
 * The second one is the real malware which implements the functions listed below.

## Functions
 * [Certificate pinning with double couple keys] (#certificate-pinning-with-double-couple-keys-notes)
 * Proxy (ssl tunneling?)
 * POST and GET for files and directories (sshfs)
 * Keylogger
 * Grab saved password in the browser
 * Grab LUKS master key (with [t4ub](https://github.com/StayPirate/t4ub))
 * Reverse shell
 * Take screenshot
 * Connection to C2s (Command and Control Server)
  > C2s as hidden service (connect zombies through [TOR](https://gitweb.torproject.org/tor.git/))

 * Disinstallation

Each functions should works on each target OSes.

## Target OSs
 1. Windows 7/8/8.1/10
 2. GNU/Linux
 3. Mac OSX

### Certificate pinning with double certificate
The bootmaster have to create his own CA, that can be done by [gen-keys.go](gen-keys.go), at first setup.

The certificate of this CA will be encoded into clients. That ensure us which zombies will connect only with certified C2S.

#### Private C2s's key
To deploy a Command and Control Server, we have first generate it's own certificate then sing it by CA

Of course, when we are going to deploy a C2S we have to deploy too: the private key and relative certificate with a valid signature.

Sometimes we want to deploy the C2S into a *not our* VPS (maybe owned by someone who doesn't know that we are using his server)... then the problem is:
> what if the real owner find our C2S and keeps a backup of keys?

He could use/abuse/steal our botnet. That's why we have a *couple of certificate* encoded in each clients.

#### Bootmaster (BM) keys
At first setup, using [gen-keys.go](gen-keys.go), there will be generated BM pub and pvt keys.
> *Note*: The pub key is not a x509 cert, but just a pub key. Like ssh-keys.

The pubkey will be encoded into clients as happens for the CA's certificate. Meanwhile the pvtkey, should remain in the BM's computer.

The pvtkey will be used to sign every command the BM wants to send to the zombies. The zombies will check the signature before execute each command they receive.

A copy of the BM's pubkey will be encoded into the C2S too, hence it can decide if forward the command incoming to the zombies by checking the signature first.

But there is a problem here:
> What if the real owner of the server, keeps a copy of one valid command taht we sent to the zombies, and replay it infinite times?

He could DOS our botnet or iterate attacks (DDOS) that we had already stopped.

#### One-Time-Command

....


#### Doble couple keys
Un'altra coppia di chiavi viene usata per firmare ogni comando che viene impartito, e questo viene fatto direttamente sulla macchina del botmaster, in modo che la chiave privata che firma i messaggi non debba essere caricata in rete, ma può rimanere isolata (offline) nella macchina del botmaster.

I client avranno quindi hardcodata la chiave pubblica di questa seconda coppia di chiavi, che utilizzeranno per validare la firma per ogni messaggio che gli arriva.

Ogni volta che un comando viene impartito dovrà avere una firma diversa da tutti i comandi inviati precedentemente altrimenti, se intercettato, potrebbe essere riusato all'infinto.

Per fare questo verrà generato un nonce dal botmaster ogni volta che un comando venga impartito, sarà fatto sia se si vuole comandare un solo client o un insieme di questi. Una volta generato viene comunicato a al/ai client interessati, che lo memorizzeranno.

Successivamente il botmaster appende al comando che vuole impartire il nonce, in una forma del tipo: comando+"\n"+nonce (dovrà essere abbastanza lungo >=128 chars).

Così facendo questo comando firmato sarà valido solo per una voltà, ovvero un **one-time-command**... in quanto se si volesse rimpartire lo stesso comando una seconda volta, bisognerebbe prima generare un nuovo nonce, ricondividerlo e quindi generare una firma valida.

I client non accettano che lo stesso nonce sia utilizzato per impartire due comandi consecutivi.

### Certificate Revocation

Grazie al sistema della doppia coppia di chiavi, se venisse rubata la chiave privata del server ci potremmo comunque collegare ai clients in quanto il certificato sarebbe ancora funzionante (se chi ha violato il server non lo abbia cancellato, **Per questo è importante avere un backup del certificato e della chiave privata del server in produzione**), e impartire come ordine di inserire quella chiave nella lista dei certificati revocati (anchessi hardcodati nella backdoor). Quindi creare un nuovo certificato per il nuovo server, firmarlo con la nostra CA ed eventualmente impartire l'ordine ai clients di connetteri ad un nuovo indirizzo ip.


## Notes
### Socket Client/Server
 * Far comunicare client e server via socket

### Steps with RSA
 1. Generare chiavi rsa.GenerateKey(rand.Reader, 2048)
 2. Codificarle in PEM per salvarle

```golang
import (
    "crypto/x509"
    "encoding/pem"
  )

x509.MarshalPKCS1PrivateKey(key *rsa.PrivateKey) converts a private key to ASN.1 DER encoded form.
type Block struct
    Bytes   []byte // The decoded bytes of the contents. Typically a DER encoded ASN.1 structure.
pem.Encode(out io.Writer, b *Block)
```
 3. Dare la pubkey al client e la pvtkey al server
 4. Server: scrive un messaggio, lo firma e lo invia al client
 5. Client: controlla la firma, se è valida stampa il messaggio
