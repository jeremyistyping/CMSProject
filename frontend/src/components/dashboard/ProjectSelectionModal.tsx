'use client';
import React, { useState, useEffect } from 'react';
import {
    Modal,
    ModalOverlay,
    ModalContent,
    ModalHeader,
    ModalBody,
    ModalCloseButton,
    VStack,
    Box,
    Text,
    Button,
    HStack,
    Badge,
    Icon,
    Input,
    InputGroup,
    InputLeftElement,
    Spinner,
    Flex,
    useColorModeValue,
} from '@chakra-ui/react';
import { FiFolder, FiSearch, FiCalendar, FiTrendingUp } from 'react-icons/fi';
import { Project } from '@/types/project';

interface ProjectSelectionModalProps {
    isOpen: boolean;
    onClose: () => void;
    projects: Project[];
    loading?: boolean;
    onSelectProject: (project: Project) => void;
}

export const ProjectSelectionModal: React.FC<ProjectSelectionModalProps> = ({
    isOpen,
    onClose,
    projects,
    loading = false,
    onSelectProject,
}) => {
    const [searchQuery, setSearchQuery] = useState('');
    const [filteredProjects, setFilteredProjects] = useState<Project[]>(projects);

    const cardBg = useColorModeValue('white', 'gray.800');
    const cardHoverBg = useColorModeValue('gray.50', 'gray.700');
    const borderColor = useColorModeValue('gray.200', 'gray.600');

    useEffect(() => {
        if (!searchQuery.trim()) {
            setFilteredProjects(projects);
        } else {
            const query = searchQuery.toLowerCase();
            const filtered = projects.filter(
                (project) =>
                    project.project_name?.toLowerCase().includes(query) ||
                    project.project_code?.toLowerCase().includes(query) ||
                    project.location?.toLowerCase().includes(query)
            );
            setFilteredProjects(filtered);
        }
    }, [searchQuery, projects]);

    const getStatusColor = (status: string) => {
        switch (status?.toUpperCase()) {
            case 'ONGOING':
                return 'green';
            case 'PLANNING':
                return 'blue';
            case 'COMPLETED':
                return 'gray';
            case 'ON_HOLD':
                return 'orange';
            default:
                return 'gray';
        }
    };

    const formatStatus = (status: string) => {
        switch (status?.toUpperCase()) {
            case 'ONGOING':
                return 'Berlangsung';
            case 'PLANNING':
                return 'Perencanaan';
            case 'COMPLETED':
                return 'Selesai';
            case 'ON_HOLD':
                return 'Ditunda';
            default:
                return status;
        }
    };

    return (
        <Modal isOpen={isOpen} onClose={onClose} size="xl" isCentered>
            <ModalOverlay backdropFilter="blur(4px)" />
            <ModalContent maxH="90vh">
                <ModalHeader>Pilih Proyek untuk Daily Report</ModalHeader>
                <ModalCloseButton />
                <ModalBody pb={6}>
                    {/* Search Input */}
                    <InputGroup mb={4}>
                        <InputLeftElement pointerEvents="none">
                            <Icon as={FiSearch} color="gray.400" />
                        </InputLeftElement>
                        <Input
                            placeholder="Cari proyek berdasarkan nama, kode, atau lokasi..."
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                            autoFocus
                        />
                    </InputGroup>

                    {/* Loading State */}
                    {loading ? (
                        <Flex justify="center" align="center" minH="200px">
                            <VStack spacing={3}>
                                <Spinner size="lg" color="brand.500" thickness="3px" />
                                <Text color="gray.500">Memuat proyek...</Text>
                            </VStack>
                        </Flex>
                    ) : (
                        <VStack spacing={3} align="stretch" maxH="500px" overflowY="auto">
                            {filteredProjects.length === 0 ? (
                                <Box
                                    textAlign="center"
                                    py={10}
                                    px={6}
                                    bg={cardBg}
                                    borderRadius="lg"
                                    borderWidth="2px"
                                    borderStyle="dashed"
                                    borderColor={borderColor}
                                >
                                    <Icon as={FiFolder} boxSize={12} color="gray.400" mb={4} />
                                    <Text color="gray.600" fontSize="lg" fontWeight="medium">
                                        {searchQuery
                                            ? 'Tidak ada proyek yang sesuai dengan pencarian'
                                            : 'Tidak ada proyek berlangsung'}
                                    </Text>
                                    {!searchQuery && (
                                        <Text color="gray.500" fontSize="sm" mt={2}>
                                            Belum ada proyek yang sedang berlangsung
                                        </Text>
                                    )}
                                </Box>
                            ) : (
                                filteredProjects.map((project) => (
                                    <Box
                                        key={project.id}
                                        p={4}
                                        bg={cardBg}
                                        borderWidth="1px"
                                        borderColor={borderColor}
                                        borderRadius="lg"
                                        cursor="pointer"
                                        transition="all 0.2s"
                                        _hover={{
                                            bg: cardHoverBg,
                                            transform: 'translateY(-2px)',
                                            shadow: 'md',
                                        }}
                                        onClick={() => {
                                            onSelectProject(project);
                                            onClose();
                                        }}
                                    >
                                        <HStack justify="space-between" align="start">
                                            <VStack align="start" spacing={2} flex={1}>
                                                <HStack spacing={2}>
                                                    <Icon as={FiFolder} color="brand.500" />
                                                    <Text fontWeight="bold" fontSize="md">
                                                        {project.project_name}
                                                    </Text>
                                                </HStack>

                                                <HStack spacing={3} fontSize="sm" color="gray.600">
                                                    <HStack spacing={1}>
                                                        <Text fontWeight="medium">Kode:</Text>
                                                        <Text>{project.project_code || '-'}</Text>
                                                    </HStack>
                                                    {project.location && (
                                                        <HStack spacing={1}>
                                                            <Text>•</Text>
                                                            <Text>{project.location}</Text>
                                                        </HStack>
                                                    )}
                                                </HStack>

                                                {project.progress !== undefined && (
                                                    <HStack spacing={2} fontSize="sm">
                                                        <Icon as={FiTrendingUp} color="green.500" boxSize={4} />
                                                        <Text color="gray.600">
                                                            Progress: <strong>{project.progress}%</strong>
                                                        </Text>
                                                    </HStack>
                                                )}
                                            </VStack>

                                            <Badge
                                                colorScheme={getStatusColor(project.status)}
                                                fontSize="xs"
                                                px={2}
                                                py={1}
                                                borderRadius="md"
                                            >
                                                {formatStatus(project.status)}
                                            </Badge>
                                        </HStack>
                                    </Box>
                                ))
                            )}
                        </VStack>
                    )}

                    {/* Footer Help Text */}
                    {!loading && filteredProjects.length > 0 && (
                        <Text fontSize="xs" color="gray.500" mt={4} textAlign="center">
                            Klik proyek untuk melanjutkan ke form daily report
                        </Text>
                    )}
                </ModalBody>
            </ModalContent>
        </Modal>
    );
};

export default ProjectSelectionModal;
