# Protocolo Principal 
O tamanho máximo do pacote será 1472 bytes, com o objetivo de atingir o MTU(Maximum Transmission Unit) do protocolo UDP. Este tamanho também será usado para todos os pacotes TCP. 
Dentro do primeiro byte iremos ter 2 bits que indicam 1 de 4 possiveis subtipos de messagem: UDP Stream, TCP Metadata, TCP Flood e TCP Keep Alive.
#### UDP Stream
Dentro do primeiro byte deste protocolo especifico iremos ter um bool que indica se a mensagem será reencaminhada para outro destinátario ou se é para o nó que está a receber.
Iremos usar 4 bytes para indicar o tamanho do payload, 4 bytes para o número do segmento e outros 4 bytes para indicar o destino do pacote.
O tamanho alocado para o payload serão 1458 bytes.
#### TCP Metadata
#### TCP Flood
1 byte para o "próximo"
1 byte para indicar o destino
#### TCP Keep Alive
1 byte para o "próximo"
1 byte para indicar o destino
