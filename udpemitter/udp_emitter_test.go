package udpemitter_test

import (
	"github.com/cloudfoundry-incubator/dropsonde-common/emitter"
	"github.com/cloudfoundry-incubator/dropsonde/udpemitter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net"
)

var _ = Describe("UdpEmitter", func() {
	var testData = []byte("hello")

	Describe("Close()", func() {
		It("closes the UDP connection", func() {

			udpEmitter, _ := udpemitter.NewUdpEmitter("localhost:42420")

			udpEmitter.Close()

			err := udpEmitter.Emit(testData)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("use of closed network connection"))
		})
	})

	Describe("Emit()", func() {
		var udpEmitter emitter.ByteEmitter

		Context("when the agent is listening", func() {

			var agentListener net.PacketConn

			BeforeEach(func() {
				var err error
				agentListener, err = net.ListenPacket("udp4", "")
				Expect(err).ToNot(HaveOccurred())

				udpEmitter, err = udpemitter.NewUdpEmitter(agentListener.LocalAddr().String())
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				agentListener.Close()
			})

			It("should send the data", func(done Done) {
				err := udpEmitter.Emit(testData)
				Expect(err).ToNot(HaveOccurred())

				buffer := make([]byte, 4096)
				readCount, _, err := agentListener.ReadFrom(buffer)
				Expect(err).ToNot(HaveOccurred())
				Expect(buffer[:readCount]).To(Equal(testData))

				close(done)
			})
		})

		Context("when the agent is not listening", func() {
			BeforeEach(func() {
				udpEmitter, _ = udpemitter.NewUdpEmitter("localhost:12345")
			})

			It("should attempt to send the data", func() {
				err := udpEmitter.Emit(testData)
				Expect(err).ToNot(HaveOccurred())
			})

			Context("then the agent starts Listening", func() {
				It("should eventually send data", func(done Done) {
					err := udpEmitter.Emit(testData)
					Expect(err).ToNot(HaveOccurred())

					agentListener, err := net.ListenPacket("udp4", ":12345")
					Expect(err).ToNot(HaveOccurred())

					err = udpEmitter.Emit(testData)
					Expect(err).ToNot(HaveOccurred())

					buffer := make([]byte, 4096)
					readCount, _, err := agentListener.ReadFrom(buffer)
					Expect(err).ToNot(HaveOccurred())
					Expect(buffer[:readCount]).To(Equal(testData))

					close(done)
				})
			})
		})
	})

	Describe("NewUdpEmitter()", func() {
		Context("when ResolveUDPAddr fails", func() {
			It("returns an error", func() {
				emitter, err := udpemitter.NewUdpEmitter("invalid-address:")
				Expect(emitter).To(BeNil())
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when all is good", func() {
			It("creates an emitter", func() {
				emitter, err := udpemitter.NewUdpEmitter("localhost:123")
				Expect(emitter).ToNot(BeNil())
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
